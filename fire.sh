#!/usr/bin/env bash

#cd `dirname $0`
set -x
cd ~/git/order-gatling

bin='dist/order-gatling'


if [ -z "$bin" ]; then
    echo "order-gatling binary not found in directory dist/" 1>&2
    exit 1
fi

echo "Binary: $bin"

platform=local

partyids="--copy-credentials-fom-config --party-id YBT --party-id-source national_id_of_natural_person --party-role-qualifier natural_person --party-role investment_decision_maker"
partyids=""
symbols=(MONCCB_USD MONHOP_USD MONHUN_USD MONLAZ_USD MONLIA_USD MONLIL_USD MONPAR_USD MONPAV_USD MONPOP_USD MONROU_USD MONSTK_USD MONSUN_USD)
refprices=(100.00 99.98 100.00 100.00 100.00 100.05 100.07 100.06 99.98 99.96 100.26 100)
accounts=(a3233761-375d-4792-8ad8-2fb9dc2080ca abf519df-d9b3-486e-8b8f-be156bf26d47 1dc5acf2-57e2-4ca4-a786-dea64dfb0f98 d52bd888-d42a-4df7-89f6-1c862d779cde dcfdd6dc-734a-45d3-8202-ad8f27a64247 72714799-9d5e-4672-af70-8b2b2be83717 d5b91838-c35a-4085-aab1-e4830df99e2d c7fd815e-6713-41a5-9df0-be0f0c6b266a b2db4f20-baef-4fc0-b994-eaea09bfaf5a 3d64fa66-b70b-40b8-96ca-a8bfd4ac52da)
#accounts=(abf519df-d9b3-486e-8b8f-be156bf26d47 1dc5acf2-57e2-4ca4-a786-dea64dfb0f98 d52bd888-d42a-4df7-89f6-1c862d779cde dcfdd6dc-734a-45d3-8202-ad8f27a64247 72714799-9d5e-4672-af70-8b2b2be83717 d5b91838-c35a-4085-aab1-e4830df99e2d c7fd815e-6713-41a5-9df0-be0f0c6b266a b2db4f20-baef-4fc0-b994-eaea09bfaf5a 3d64fa66-b70b-40b8-96ca-a8bfd4ac52da)
if [ $platform = "test" ]; then
    session_prefix=test-exactpro
elif [ $platform = "dev" ]; then
    session_prefix=dev-exactpro
elif [ $platform = "local" ]; then
    session_prefix=bff-exactpro
else
    echo "Platform not handled: $platform" >&2
    exit 1
fi

allsymbols=$(IFS=, ; echo "${symbols[*]}")
allrefprices=$(IFS=, ; echo "${refprices[*]}")

mode=stable-rate

if [ $mode != "stable-rate" ]; then
    for i in `seq 0 1 9`; do
        $bin \
            --context ${session_prefix}$((i+1)) \
            --metrics --port $((5010+$i+1)) \
            --symbols ${symbols[$i]} \
            --refprices ${refprices[$i]} \
            --accounts ${accounts[$i]} \
            --update-tempo 0s \
            --no-mass-cancel \
            $partyids \
            -v &
    done
else
    for i in `seq 0 1 9`; do
        $bin \
            --context ${session_prefix}$((i+1)) \
            --metrics --port $((5010+$i+1)) \
            --symbols $allsymbols \
            --refprices $allrefprices \
            --accounts ${accounts[$i]} \
            --order-rate 10 \
            --no-mass-cancel \
            $partyids \
            -v &
    done
fi

