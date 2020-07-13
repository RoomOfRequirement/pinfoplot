#

## pinfoplot

thanks to [gopsutil](https://github.com/shirou/gopsutil) and [gonum/plot](https://github.com/gonum/plot)

notice: this tool only works on `Linux` (according to `gopsutil`)

```sh
pinfoplot tool in golang to generate process info image
Version: 0.0.1
Usage: pinfoplot [-help help] [-v version] [-p pid] [-d sampling duration] [-i sampling interval] [-w output image width (cm or inch)] [-h output image height (cm or inch)] [-o output image file path]
Options
  -d duration
        sampling duration (0 means sample until pid exits) (default 10s)
  -h string
        output image height (cm or inch) (default "8cm")
  -help
        help info
  -i duration
        sampling interval (default 50ms)
  -o string
        output image file path (default "pinfo.png")
  -p int
        pid to get info from (default -1)
  -v    version info
  -w string
        output image width (cm or inch) (default "10cm")
```