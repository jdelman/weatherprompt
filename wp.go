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
)

// structs for unpacking JSON
type Conditions struct {
  Current_observation Current
}

type Current struct {
  Station_id           string
  Weather              string
}

type CachedConditions struct {
  Last                 int64   `json:"last"`
  Station              string  `json:"station"`
  Condition            string  `json:"condition"`
  Emoji                string  `json:"emoji"`
}

type Zipcode struct {
  Postal               string
}

var (
  debug                bool
  api_key              string
  wait_minutes         int64
  user_zip             string
  force_check          bool
)

const WAIT_MINUTES_DEFAULT = 10

// declare EMOJI map.
func GetEmoji() map[string]string {
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

// Fetch does URL processing
func Fetch(url string) ([]byte, error) {
  res, err := http.Get(url)
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


func SaveCurrentConditions(station string, condition string, emoji string) {
  var cond CachedConditions

  cond.Station = station
  cond.Condition = condition
  cond.Emoji = emoji
  cond.Last = time.Now().Unix()

  b, err := json.Marshal(cond)
  CheckError(err, "Marshall SaveCurrentConditions")

  err2 := ioutil.WriteFile((os.Getenv("HOME") + "/.current_conditions"), b, 0644)  
  CheckError(err2, "WriteFile SaveCurrentConditions")
}


func MapConditionToEmoji(condition string) string {
  for cond, emoji := range GetEmoji() {
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


func WeatherUrlForZip(zip string) string {
  const stem = "http://api.wunderground.com/api/"
  const prezip = "/conditions/q/zmw:"
  const postzip = ".1.99999.json"

  fullUrl := stem + api_key + prezip + zip + postzip
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
  flag.Parse()
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
    fmt.Println(cachedcond.Emoji)
    os.Exit(0)
  }

  if user_zip == "" {
    DebugPrint("zip lookup")
    user_zip = GetZip()
  } else {
    DebugPrint("using forced zip:", user_zip)
  }

  url := WeatherUrlForZip(user_zip)
  b, err := Fetch(url)
  CheckError(err, "Fetch Weather")
  DebugPrint("got conditions:", string(b[:]))

  var cond Conditions
  err = json.Unmarshal(b, &cond)
  CheckError(err, "Unmarshall conditions JSON")

  station := cond.Current_observation.Station_id
  condition := cond.Current_observation.Weather
  emoji := MapConditionToEmoji(cond.Current_observation.Weather)

  // save conditions
  SaveCurrentConditions(station, condition, emoji)

  fmt.Println(emoji)
}