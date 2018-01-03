package main

import (
  "encoding/json"
  "flag"
  "fmt"
  "io/ioutil"
  "net/http"
  "os"
  "strings"
  "time"
  "strconv"
)

// structs for unpacking JSON
type Astronomy struct {
  Moon_phase           Moon
  Sun_phase            Sun
}

type Moon struct {
  PhaseofMoon          string
}

type Sun struct {
  Sunrise              SmallTime
  Sunset               SmallTime
}

type SmallTime struct {
  Hour                 string
  Minute               string
}

type Conditions struct {
  Current_observation Current
}

type Current struct {
  Station_id           string
  Weather              string
  Temp_f               float64
}

type CachedConditions struct {
  Last                 int64   `json:"last"`
  Station              string  `json:"station"`
  Condition            string  `json:"condition"`
  Emoji                string  `json:"emoji"`
  MoonEmoji            string  `json:"moon_emoji"`
  Temp                 string  `json:"temp"`
}

type Zipcode struct {
  Postal               string
}

// command line flags
var (
  debug                bool
  api_key              string
  wait_minutes         int64
  user_zip             string
  force_check          bool
  show_moon            bool
  show_temp            bool
)

const WAIT_MINUTES_DEFAULT = 10

// declare EMOJI map.
func GetWeatherEmoji() map[string]string {
  EMOJI := map[string]string {
    "Drizzle": "🌦",
    "Rain": "☔",
    "Snow": "🌨",
    "Snow Grains": "🌨",
    "Ice Crystals": "🌨",
    "Ice Pellets": "🌨",
    "Hail": "🌧",
    "Mist": "🌫",
    "Fog": "🌫",
    "Fog Patches": "🌫",
    "Smoke": "🌪",
    "Volcanic Ash": "🌪",
    "Widespread Dust": "🏜",
    "Sand": "🏜",
    "Haze": "🌫",
    "Spray": "🌦",
    "Dust Whirls": "🏜",
    "Sandstorm": "🏜",
    "Low Drifting Snow": "🌨",
    "Low Drifting Widespread Dust": "🏜",
    "Low Drifting Sand": "🏜",
    "Blowing Snow": "🌬❄",
    "Blowing Widespread Dust": "🌬🏜",
    "Blowing Sand": "🌬🏜",
    "Rain Mist": "🌦",
    "Rain Showers": "☔",
    "Snow Showers": "🌨",
    "Snow Blowing Snow Mist": "🌬🌨",
    "Ice Pellet Showers": "🌨☄",
    "Hail Showers": "🌧",
    "Small Hail Showers": "🌧",
    "Thunderstorm": "🌩",
    "Thunderstorms and Rain": "⛈",
    "Thunderstorms and Snow": "🌩🌨",
    "Thunderstorms and Ice Pellets": "🌩☄",
    "Thunderstorms with Hail": "⛈",
    "Thunderstorms with Small Hail": "⛈",
    "Freezing Drizzle": "🌨",
    "Freezing Rain": "🌨",
    "Freezing Fog": "🌫",
    "Patches of Fog": "🌫",
    "Shallow Fog": "🌫",
    "Partial Fog": "🌫",
    "Overcast": "☁",
    "Clear": "🌞",
    "Partly Cloudy": "🌤",
    "Mostly Cloudy": "🌥",
    "Scattered Clouds": "⛅",
    "Small Hail": "🌧",
    "Squalls": "🌊",
    "Funnel Cloud": "🌪",
    "Unknown Precipitation": "🌧❔",
    "Unknown": "❔",
  }
  return EMOJI
}

func GetMoonEmoji() map[string]string {
  MOON_EMOJI := map[string]string {
    "New": "🌚",
    "Waxing Crescent": "🌙",
    "First Quarter": "🌛",
    "Waxing Gibbous": "🌔",
    "Full": "🌝",
    "Waning Gibbous": "🌖",
    "Last Quarter": "🌜",
    "Waning Crescent": "🌘",
  }
  return MOON_EMOJI
}

// Fetch does URL processing
func Fetch(url string) ([]byte, error) {
  timeout := time.Duration(3 * time.Second)
  client := http.Client{
    Timeout: timeout,
  }
  res, err := client.Get(url)
  CheckError(err, "fetch")
  if res.StatusCode != 200 {
    fmt.Fprintf(os.Stderr, "Bad HTTP Status: %d\n", res.StatusCode)
    return nil, err
  }
  b, err := ioutil.ReadAll(res.Body)
  res.Body.Close()
  return b, err
}


func GetCachedConditions() CachedConditions {
  var conditions CachedConditions

  if b, err := ioutil.ReadFile(os.Getenv("HOME") + "/.current_conditions"); err == nil {
    jsonErr := json.Unmarshal(b, &conditions)
    CheckError(jsonErr, "GetCachedConditions")
  }

  return conditions
}


func SaveCurrentConditions(station string, condition string, emoji string, moonemoji string, temp string) {
  var cond CachedConditions

  cond.Station = station
  cond.Condition = condition
  cond.Emoji = emoji
  cond.MoonEmoji = moonemoji
  cond.Temp = temp
  cond.Last = time.Now().Unix()

  b, err := json.Marshal(cond)
  CheckError(err, "Marshall SaveCurrentConditions")

  err2 := ioutil.WriteFile((os.Getenv("HOME") + "/.current_conditions"), b, 0644)  
  CheckError(err2, "WriteFile SaveCurrentConditions")
}


func MapMoonPhaseToEmoji(phase string) string {
  for phs, emoji := range GetMoonEmoji() {
    if strings.HasPrefix(phase, phs) {
      return emoji
    }
  }
  return ""
}


func MapConditionToEmoji(condition string) string {
  for cond, emoji := range GetWeatherEmoji() {
    if strings.HasSuffix(condition, cond) {
      return emoji
    }
  }
  return ""
}


func GetZip() string {
  b, err := Fetch("http://ipinfo.io/json")
  CheckError(err, "fetch->getzip")

  // fmt.Println("%s", string(b[:]))

  var zipcode Zipcode
  jsonErr := json.Unmarshal(b, &zipcode)
  CheckError(jsonErr, "zipcode json")

  return zipcode.Postal
}


// CheckError exits on error with a message
func CheckError(err error, tag string) {
  if err != nil {
    fmt.Fprintf(os.Stderr, "%s: Fatal error\n%v\n", tag, err)
    os.Exit(1)
  }
}


func DebugPrint(a ...interface{}) {
  if debug {
    fmt.Println(a...)
  }
}


func WeatherUrlForZip(section string, zip string) string {
  const stem = "http://api.wunderground.com/api/"
  const prezip = "/q/zmw:"
  const postzip = ".1.99999.json"

  fullUrl := stem + api_key + "/" + section + prezip + zip + postzip
  DebugPrint("using weather url:", fullUrl)

  return fullUrl
}


func TimeToCheckYet(timeWas int64) bool {
  ok := time.Now().Unix() > timeWas + (60 * wait_minutes)
  DebugPrint("TimeToCheckYet?", ok)
  return ok
}


func ParseCommandLine() {
  flag.Int64Var(&wait_minutes, "w", WAIT_MINUTES_DEFAULT, "Number of minutes to wait before checking")
  flag.BoolVar(&debug, "d", false, "Turn on debug mode")
  flag.StringVar(&api_key, "k", "", "API key for api.wunderground.com")
  flag.StringVar(&user_zip, "z", "", "Force zip code (skip ipinfo.io lookup)")
  flag.BoolVar(&force_check, "f", false, "Force lookup (don't use cached data)")
  flag.BoolVar(&show_moon, "m", false, "Include the phase of the moon (at night)")
  flag.BoolVar(&show_temp, "t", false, "Show the temperature in Fahrenheit")
  flag.Parse()
}


func WithHourAndMinute(t time.Time, smalltime SmallTime) time.Time {
  hour, hErr := strconv.Atoi(smalltime.Hour)
  CheckError(hErr, "hour conversion to string")

  minute, mErr := strconv.Atoi(smalltime.Minute)
  CheckError(mErr, "minute conversion to string")

  return time.Date(t.Year(), t.Month(), t.Day(),
                   hour, minute,
                   t.Second(), t.Nanosecond(), t.Location())
}


func main() {
  ParseCommandLine()

  DebugPrint("Debug mode ON")

  var cachedcond CachedConditions
  cachedcond = GetCachedConditions()
  DebugPrint("last check was", cachedcond.Last)
  if !force_check && cachedcond.Last != 0 && !TimeToCheckYet(cachedcond.Last) {
    // return cached; we're done!
    DebugPrint("using cached response")
    out := cachedcond.Emoji;
    if show_moon {
      out += " " + cachedcond.MoonEmoji
    }
    if show_temp {
      out += "  " + cachedcond.Temp + "°"
    }
    fmt.Println(out)
    os.Exit(0)
  }

  if user_zip == "" {
    DebugPrint("zip lookup")
    user_zip = GetZip()
  } else {
    DebugPrint("using forced zip:", user_zip)
  }

  url := WeatherUrlForZip("conditions", user_zip)
  DebugPrint(url)
  b, err := Fetch(url)
  CheckError(err, "Fetch Weather")
  DebugPrint("got conditions:", string(b[:]))

  var cond Conditions
  err = json.Unmarshal(b, &cond)
  CheckError(err, "Unmarshall conditions JSON")

  moonemoji := ""
  if show_moon {
    url = WeatherUrlForZip("astronomy", user_zip)
    b, err = Fetch(url)
    CheckError(err, "Fetch Astronomy")
    DebugPrint("astronomy url:", url, "got astronomy:", string(b[:]))

    var astronomy Astronomy
    err = json.Unmarshal(b, &astronomy)
    CheckError(err, "Unmarshall astronomy JSON")
    DebugPrint("astronomy", astronomy)

    // use sunrise/sunset to decide if we should really show the moon
    DebugPrint("sunrise", astronomy.Sun_phase.Sunrise, "sunset", astronomy.Sun_phase.Sunset)
    // sunrise := WithHourAndMinute(time.Now(), astronomy.Sun_phase.Sunrise)
    sunset := WithHourAndMinute(time.Now(), astronomy.Sun_phase.Sunset)
    // DebugPrint("sunrise", sunrise, "sunset", sunset)

    if time.Now().After(sunset) {
      DebugPrint("it's night")
      moonemoji = MapMoonPhaseToEmoji(astronomy.Moon_phase.PhaseofMoon)
    } else {
      moonemoji = ""
    }
  }

  station := cond.Current_observation.Station_id
  condition := cond.Current_observation.Weather
  temp := strconv.FormatFloat(cond.Current_observation.Temp_f, 'f', 0, 64);
  emoji := MapConditionToEmoji(cond.Current_observation.Weather)

  // save conditions
  SaveCurrentConditions(station, condition, emoji, moonemoji, temp)

  out := emoji
  if show_moon {
    out += " " + moonemoji
  }
  if show_temp {
    out += "  " + temp + "°"
  }

  fmt.Println(out)
}