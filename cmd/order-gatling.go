package cmd

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"sylr.dev/fix/config"
	"sylr.dev/fix/pkg/initiator"
	"sylr.dev/fix/pkg/utils"

	"github.com/alexppxela/order-gatling/order"
)

var Version = "dev"

var (
	optionSymbols       []string
	optionRefPrices     []float64
	optionAccounts      []string
	optionUpdateTempo   time.Duration
	optionNoMassCancel  bool
	optionQuoteWorkflow bool
	optionNewOrderRate  uint
)

// OrderGatlingCmd represents the base command when called without any subcommands.
var OrderGatlingCmd = &cobra.Command{
	Use:          "order-gatling",
	SilenceUsage: true,
	Version:      Version,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := initiator.ValidateOptions(cmd, args); err != nil {
			return err
		}

		if err := validate(); err != nil {
			return err
		}

		if err := InitHTTP(); err != nil {
			return err
		}

		return InitLogger()
	},
	RunE: execute,
}

func init() {
	options := config.GetOptions()
	configPath := strings.Join([]string{"$HOME", ".fix", "config"}, string(os.PathSeparator))

	OrderGatlingCmd.PersistentFlags().StringVar(&options.Config, "config", os.ExpandEnv(configPath), "Config file")
	OrderGatlingCmd.PersistentFlags().CountVarP(&options.Verbose, "verbose", "v", "Increase verbosity")
	OrderGatlingCmd.PersistentFlags().BoolVar(&options.LogCaller, "log-caller", false, "Add caller info to log lines")
	OrderGatlingCmd.PersistentFlags().BoolVar(&options.Interactive, "interactive", true, "Enable interactive mode")
	OrderGatlingCmd.PersistentFlags().BoolVar(&options.Metrics, "metrics", false, "Enable metrics")
	OrderGatlingCmd.PersistentFlags().BoolVar(&options.PProf, "pprof", false, "Enable pprof")
	OrderGatlingCmd.PersistentFlags().IntVar(&options.HTTPPort, "port", 5009, "HTTP port")
	OrderGatlingCmd.PersistentFlags().BoolP("help", "h", false, "Help for fix")
	OrderGatlingCmd.PersistentFlags().Bool("version", false, "Version for fix")

	OrderGatlingCmd.PersistentFlags().StringSliceVar(&optionSymbols, "symbols", nil, "Symbols")
	OrderGatlingCmd.PersistentFlags().Float64SliceVar(&optionRefPrices, "refprices", nil, "Reference price")
	OrderGatlingCmd.PersistentFlags().StringSliceVar(&optionAccounts, "accounts", nil, "Accounts sent in PartyIDs")
	OrderGatlingCmd.PersistentFlags().DurationVar(&optionUpdateTempo, "update-tempo", 0*time.Millisecond, "Duration before updating order (ms)")
	OrderGatlingCmd.PersistentFlags().BoolVar(&optionNoMassCancel, "no-mass-cancel", false, "Do not send mass order cancel request")
	OrderGatlingCmd.PersistentFlags().BoolVar(&optionQuoteWorkflow, "quote", false, "Use quote instead of order workflow")
	OrderGatlingCmd.PersistentFlags().UintVar(&optionNewOrderRate, "order-rate", 0, "Number of new order sent per second")

	initiator.AddPersistentFlags(OrderGatlingCmd)
	_ = initiator.AddPersistentFlagCompletions(OrderGatlingCmd)

	_ = OrderGatlingCmd.MarkFlagRequired("symbols")

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	if err := viper.BindPFlags(OrderGatlingCmd.PersistentFlags()); err != nil {
		panic(err)
	}
}

func validate() error {
	if len(optionSymbols) == 0 {
		return errors.New("missing symbol list")
	}
	if len(optionSymbols) != len(optionRefPrices) {
		return errors.New("number of symbols must match number of reference prices")
	}
	if len(optionAccounts) == 0 {
		return errors.New("missing account list")
	}
	return nil
}

func execute(cmd *cobra.Command, args []string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, os.Kill)

	orderSender, err := createOrderSender(ctx)
	if err != nil {
		cancel()
		return err
	}
	err = orderSender.Connect()
	if err != nil {
		cancel()
		return err
	}

	if optionNewOrderRate == 0 {
		orderManager := order.NewManager(
			ctx,
			orderSender,
			optionAccounts,
			optionSymbols,
			optionRefPrices,
			optionQuoteWorkflow,
			optionUpdateTempo)

		if !optionNoMassCancel {
			orderManager.CancelAllOrders()
			<-time.After(2 * time.Second)
		}
		orderManager.Start()
	} else {
		orderManager := order.NewSampledManager(
			ctx,
			orderSender,
			optionAccounts,
			optionSymbols,
			optionRefPrices,
			optionNewOrderRate)

		if !optionNoMassCancel {
			orderManager.CancelAllOrders()
			<-time.After(2 * time.Second)
		}
		orderManager.Start()
	}

	<-ctx.Done()
	config.GetLogger().Info().Msg("Received signal. Stopping services")
	cancel()
	<-orderSender.Closed
	config.GetLogger().Trace().Msg("orderSender is closed")

	return nil
}

func createOrderSender(ctx context.Context) (*order.SenderApp, error) {
	configContext, err := config.GetCurrentContext()
	if err != nil {
		return nil, err
	}

	sessions, err := configContext.GetSessions()
	if err != nil {
		return nil, err
	}

	session := sessions[0]
	transportDict, appDict, err := session.GetFIXDictionaries()
	if err != nil {
		return nil, err
	}

	settings, err := configContext.ToQuickFixInitiatorSettings()
	if err != nil {
		return nil, err
	}

	qfLogger := utils.QuickFixAppMessageLogger{Logger: config.GetLogger(), TransportDataDictionary: transportDict, AppDataDictionary: appDict}

	app, err := order.NewOrderSender(ctx, qfLogger, settings, session)
	if err != nil {
		return nil, err
	}

	return app, nil
}
