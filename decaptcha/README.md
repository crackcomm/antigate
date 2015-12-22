# decaptcha

Command-line tool for antigate.com

Looks for antigate key in `ANTIGATE_KEY` environment variable or `--antigate-key` flag.

```sh
$ go install github.com/crackcomm/antigate/decaptcha
$ decaptcha --help
Usage of decaptcha:
  -antigate-key string
    	Antigate API Key (also accepts ANTIGATE_KEY environment variable)
  -captcha-url string
    	Captcha URL
  -output-image string
    	Output image
```
