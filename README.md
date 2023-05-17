### MarketMaking animator

## Description
The function of this repository is to generate order cancel replace requests to market all the time to generate heavy load.

Code is based on [fix toolbox](https://github.com/sylr/fix)

## How does it works
At startup, tool send an [OrderMassCancelRequest](https://fiximate.fixtrading.org/en/FIX.Latest/msg50.html) for all accounts / securities / sides. Cancellation at startup can be disabled with option `--no-mass-cancel`

After awaiting, it sends one [NewOrderSingle](https://fiximate.fixtrading.org/en/FIX.Latest/msg14.html), and it updates quantity on [ExecutionReport](https://fiximate.fixtrading.org/en/FIX.Latest/msg9.html) reception.
To test [Quote](https://fiximate.fixtrading.org/en/FIX.Latest/msg27.html) workflow instead of order workflow, you can use option `--quote`

An other mode is to send [NewOrderSingle](https://fiximate.fixtrading.org/en/FIX.Latest/msg14.html) periodically. No order amendment will be done, only order creation. To activate it, option `--order-rate` must be greater than 0.

## How to build it
`make build`

## How to launch it
`dist/order-gatling`

Options are:
```
--context        : FIX context to send orders/quotes
--symbols        : List of symbol to animate
--refprices      : List of reference prices for each symbols
--accounts       : Accounts sent in PartyIDs
--metrics        : Enable metrics
--port           : HTTP port for metrics
--no-mass-cancel : Do not send mass order cancel request
--order-rate     : Number of new order sent per second
--quote          : Use quote instead of order workflow
--update-tempo   : Duration before updating order (ms)
```

### Examples
#### Order amendment 50ms after execution report acknowledge (with metrics and trace logging)
```sh
dist/order-gatling \
    --context fix-session-conf \
    --metrics --port 5011 \
    --symbols MONA_EUR,CENA_EUR \
    --refprices 101.50,100.81 \
    --accounts trader1,trader2 \
    --update-tempo 50ms \
    --no-mass-cancel \
    -vv
```

#### Quote amendment after quote status report acknowledge
```sh
dist/order-gatling \
    --context fix-session-conf \
    --symbols MONA_EUR,CENA_EUR \
    --refprices 101.50,100.81 \
    --accounts trader1,trader2 \
    --update-tempo 0ms \
    --quote
    --no-mass-cancel
```

#### Create 100 orders per second
```sh
dist/order-gatling \
    --context fix-session-conf \
    --symbols MONA_EUR,CENA_EUR \
    --refprices 101.50,100.81 \
    --accounts trader1 \
    --order-rate 100
```
