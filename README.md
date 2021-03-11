# PolkaTax

Extract stacking rewards and date time and correlate with Dot USD marker value for both KSM/DOT networks.

```console
Usage of ./polkaTax:
      --account string   The account address to use [required]
      --concurrent int   The number of concurent jobs to run (default 100)
      --csv string       To save the results to a csv file
      --era duration     Optional set the era to compute historical quote. default is 6h for Kusama and 24h for Polkadot
      --from string      Optional starting date
      --network string   The network to use [allowed: polkadot,kusama] (default "polkadot")
      --to string        Optional ending date
      --url string       The polkascan api url to use (default "https://explorer-32.polkascan.io")
```

> Liability Disclaimer. The Software is provided as is, without any representation or warranty of any kind, either express or implied, including without limitation any representations or endorsements regarding the use of, the results of, or performance of the product, its appropriateness, accuracy, reliability, or correctness. The entire risk as to the use of this product is assumed by Licensee. In no event will PolkaTax be liable for additional direct or indirect damages including any lost profits, lost savings, or other incidental or consequential damages arising from any defects, or the use or inability to use the program. In short, use it at you own risk.
