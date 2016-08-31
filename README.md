# 🌞 weatherprompt ☔
### emoji weather, moon phases, & temperature in your prompt 

Uses [ipinfo](http://ipinfo.io/json) to get your ZIP code, which is then passed to the [Wunderground API](https://api.wunderground.com/api). (You'll need a Wunderground API key - don't worry, it's free.)

Originally I wrote this in Python, but the overhead of loading Python and the imported libraries made it too slow to use on each prompt. So I re-wrote in Go, which is a compiled language, and it starts up much faster. With Go 1.7 and some compile flags, binary is a *slim* 4 MB.


#### HOWTO

1. Build: `go build -ldflags="-s -w" wp.go`

2. Make it executable: `chmod +x wp`

3. copy `wp` to a directory in your PATH.


#### Usage

`-k` is the only required flag:

`wp -k [YOUR_WUNDERGROUND_API_KEY]`

Throw it in your prompt:

`export PS1="$(wp -m -t -k [key])  \u@\h\w $ "`



#### Full list of flags

```
  -d  Turn on debug mode
  -f  Force lookup (don't use cached data)
  -k string
      API key for api.wunderground.com
  -m  Include the phase of the moon (at night)
  -t  Show the temperature in Fahrenheit
  -w int
      Number of minutes to wait before checking (default 10)
  -z string
      Force zip code (skip ipinfo.io lookup)
```

You can also type `wp -h` to see this list.

***

(C) Josh Delman, 2016