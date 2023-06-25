package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"mySongs/config"
)

var (
	recommendationsURL = "https://api.spotify.com/v1/recommendations"
	tokenURL           = "https://accounts.spotify.com/api/token"
	authorizeURL       = "https://accounts.spotify.com/authorize"
	redirectURI        = "https://localhost"
	audioURL           = "https://api.spotify.com/v1/audio-features"

	authorization = "Authorization"
	myUserID      = ""
	accessToken   = ""
)

type spotySongFeatures struct {
	Danceability     float64 `json:"danceability"`
	Energy           float64 `json:"energy"`
	Key              int     `json:"key"`
	Loudness         float64 `json:"loudness"`
	Mode             float64 `json:"mode"`
	Speechiness      float64 `json:"speechiness"`
	Acousticness     float64 `json:"acousticness"`
	Instrumentalness float64 `json:"instrumentalness"`
	Liveness         float64 `json:"liveness"`
	Valence          float64 `json:"valence"`
	Tempo            float64 `json:"tempo"`
	Type             string  `json:"type"`
	Id               string  `json:"id"`
	Uri              string  `json:"uri"`
	TrackHref        string  `json:"track_href"`
	AnalysisUrl      string  `json:"analysis_url"`
	DurationMs       float64 `json:"duration_ms"`
	TimeSignature    float64 `json:"time_signature"`
}

// danceability,energy, liveness(for soca), loudness(lower minus val is good?), speechiness(should be low too for soca? up to .33?), valence(high valence for cheerfulness!)
type songAverages struct {
	danceability float64
	energy       float64
	key          int
	liveness     float64
	loudness     float64
	instru       float64
	acousticness float64
	speechiness  float64
	valence      float64
	tempo        float64
}

type spotyAuth struct {
	aToken      string `json:"access_token"`
	tokenType   string `json:"token_type"`
	tokenExpiry string `json:"expires_in"`
}

type pageObject struct {
	Href     string        `json:"href"`
	Items    []savedObject `json:"items"`
	Limit    int           `json:"limit"`
	Next     string        `json:"next"`
	Offset   int           `json:"offset"`
	Previous string        `json:"previous"`
	Total    int           `json:"total"`
}

// Timestamp string `json:"added_at,omitempty"`
type savedObject struct {
	AddedOn string      `json:"added_at"`
	Track   trackObject `json:"track"`
}

type recommendationObject struct {
	Tracks []trackObject `json:"tracks"`
	//Seeds []seedObject      `json:"seeds"`
}

type playlistObject struct {
	Href string `json:"href"`
	ID   string `json:"id"`
	Name string `json:"name"`
	URI  string `json:"uri"`
}

type trackObject struct {
	Href string `json:"href"`
	ID   string `json:"id"`
	Name string `json:"name"`
	URI  string `json:"uri"`
}

func main() {

	//get token >> url with base64 clientID:clientSecret
	//save it and use it in AUth header for future requests!
	////have to authorize and then use second token to access user's stuff!

	//ask for features of few saved saved songs(or those that you wish to have recommendation on)
	////allow user to input songs IDs for seeding
	//calculate average of features to focus on some of the attributes--see notes

	//ask for recommendation with target features calculated above
	///cycle through songs recommendation, confirm that recommended song's feature is within range(neccessary?!? prolly not but in V2)
	////keeping only those within range of the calculated average(limited by a 10% lower/upper bound from song averages)

	///build a playlist with those suggested songs...

	cfg, err := config.FromEnvironment()
	if err != nil {
		//logger.Log("msg", "invalid configuration", "error", err)
		fmt.Printf("Invalid configuration %s", err)
		os.Exit(1)
	}

	myUserID = cfg.UserID

	client := &http.Client{}
	//sngIDs := []string{"53Uxhu7xAUjKcxcomTuMbw", "1ec8eWlZSWg4djoQaMrI1u", "5Hut81t6yks6GUAQzty2fD"}
	sngIDs := []string{"2K1zp0p7PVmrBUQu6evtfe"} //Avicii Somewhere in Stockholm

	var songFeats []spotySongFeatures

	for _, c := range sngIDs {
		res, er := getSongFeats(client, c)
		if er != nil {
			fmt.Printf("Failed: %s", er)
			return
		}
		fmt.Printf("\n\nSuccess: E:%g D:%g Live:%g Loud:%g S:%g V:%g T:%g for %s\n", res.Energy, res.Danceability, res.Liveness, res.Loudness, res.Speechiness, res.Valence, res.Tempo, c)
		songFeats = append(songFeats, res)
	}

	songAvg := calculateAverages(songFeats)
	fmt.Printf("\n\nMid: %g,%g,%g,%g,%g,%g \n", songAvg.danceability, songAvg.energy, songAvg.liveness, songAvg.loudness, songAvg.speechiness, songAvg.valence)

	recos, err := getRecommendations(songAvg, sngIDs)
	if err != nil {
		fmt.Printf("Failed: %s", err)
		return
	}

	newPlaylist, err := createNewPlaylist("aHouseTest")
	if err != nil {
		fmt.Printf("Failed: %s", err)
		return
	}
	tracks := []string{}
	for _, t := range recos.Tracks {
		tracks = append(tracks, t.URI)
	}

	addTracksToPlaylist(newPlaylist, tracks)

}

func getSongFeats(client *http.Client, songID string) (spotySongFeatures, error) {

	//https: //api.spotify.com/v1/audio-features/{id}
	rqU := fmt.Sprintf("%s/%s", audioURL, songID)
	authRequest, err := http.NewRequest("GET", rqU, nil)
	if err != nil {
		return spotySongFeatures{}, errors.New("cannot generate request")
	}

	authRequest.Header.Set(authorization, "Bearer "+accessToken)
	authRequest.Header.Add("Content-Type", "application/json")

	authResult, err := client.Do(authRequest)
	if err != nil {
		return spotySongFeatures{}, errors.New("request error")
	}

	authResponseBody, err := ReadAndCloseBody(authResult)
	if err != nil {
		return spotySongFeatures{}, errors.New("could not read response")
	}
	fmt.Printf("Got: %s", string(authResponseBody))

	var songF spotySongFeatures
	err = json.Unmarshal(authResponseBody, &songF)
	if err != nil {
		//logger.Log("err", err, "msg", fmt.Sprintf("user auth JSON error. The Response String is: %s", string(authResponseBody)))
		return spotySongFeatures{}, errors.New("decoding error")
	}

	return songF, nil
}

func getMyTracks(client *http.Client) (spotySongFeatures, error) {

	//https: //api.spotify.com/v1/audio-features/{id}
	//rqU := fmt.Sprintf("%s/%s", audioURL, songID)
	authRequest, err := http.NewRequest("GET", "https://api.spotify.com/v1/me/tracks?offset=0&limit=2", nil)
	//https://api.spotify.com/v1/me/tracks?offset=0&limit=20
	//https://api.spotify.com/v1/me/tracks
	if err != nil {
		return spotySongFeatures{}, errors.New("cannot generate request")
	}

	authRequest.Header.Set(authorization, "Bearer "+accessToken)
	authRequest.Header.Add("Content-Type", "application/json")

	authResult, err := client.Do(authRequest)
	if err != nil {
		return spotySongFeatures{}, errors.New("request error")
	}

	authResponseBody, err := ReadAndCloseBody(authResult)
	if err != nil {
		return spotySongFeatures{}, errors.New("could not read response")
	}
	fmt.Printf("Got: %s", string(authResponseBody))

	/*var songF spotySongFeatures
	err = json.Unmarshal(authResponseBody, &songF)
	if err != nil {
		//logger.Log("err", err, "msg", fmt.Sprintf("user auth JSON error. The Response String is: %s", string(authResponseBody)))
		return spotySongFeatures{}, errors.New("decoding error")
	}
	*/
	return spotySongFeatures{}, nil
}

func getSimilarTracks(mid songAverages) {
	client := &http.Client{}
	tracks, er := getMySavedTracks(client)
	if er != nil {
		fmt.Printf("Failed: %s", er)
	}
	//loop through asking features for the song_ids..should optimize it with one request..TODO

	enerLower, enerUpper := calculateRanges(mid.energy)
	danceaLower, danceaUpper := calculateRanges(mid.danceability)
	liveLower, liveUpper := calculateRanges(mid.liveness)
	loudLower, loudUpper := calculateRanges(mid.loudness)
	speechLower, speechUpper := calculateRanges(mid.speechiness)
	valLower, valUpper := calculateRanges(mid.valence)

	for _, c := range tracks.Items {
		track := c.Track
		//fmt.Printf("item: %s with ID: %s and name: %s", track.Href, track.ID, track.Name)
		res, er := getSongFeats(client, track.ID)
		if er != nil {
			fmt.Printf("Failed: %s", er)
		}
		//fmt.Printf("\nSuccess: Ener: %g , val: %g, tempo: %g for %s \n", res.Energy, res.Valence, res.Tempo, track.Name)

		sameEnergy := isInRange(res.Energy, enerLower, enerUpper)
		sameValence := isInRange(res.Valence, valLower, valUpper)
		sameDancea := isInRange(res.Danceability, danceaLower, danceaUpper)
		sameLive := isInRange(res.Liveness, liveLower, liveUpper)
		sameLoud := isInRange(res.Loudness, loudLower, loudUpper)
		sameSpeechy := isInRange(res.Speechiness, speechLower, speechUpper)
		if sameEnergy || sameValence || sameDancea || sameLive || sameLoud || sameSpeechy {
			fmt.Printf("\nFound the same vibe: %s with E:%t,D:%t,Live:%t,Loud:%t,S:%t,V:%t \n", track.Name, sameEnergy, sameDancea, sameLive, sameLoud, sameSpeechy, sameValence)
			fmt.Printf("Had>> E:%g,D:%g,Live:%g,Loud:%g,S:%g,V:%g \n", res.Energy, res.Danceability, res.Liveness, res.Loudness, res.Speechiness, res.Valence)
		}
	}
	//save those that are within range of passed parameters...
}

func getRecommendations(mid songAverages, seedTracks []string) (recommendationObject, error) {

	client := &http.Client{}

	url := fmt.Sprintf("%s?%s",
		recommendationsURL,
		url.Values{
			"limit":                   {strconv.Itoa(30)},
			"market":                  {"ES"},
			"seed_tracks":             {strings.Join(seedTracks, ",")},
			"target_energy":           {strconv.FormatFloat(mid.energy, 'f', -1, 64)},
			"target_liveness":         {strconv.FormatFloat(mid.liveness, 'f', -1, 64)},
			"target_speechiness":      {strconv.FormatFloat(mid.speechiness, 'f', -1, 64)},
			"target_acousticness":     {strconv.FormatFloat(mid.acousticness, 'f', -1, 64)},
			"target_danceability":     {strconv.FormatFloat(mid.danceability, 'f', -1, 64)},
			"target_instrumentalness": {strconv.FormatFloat(mid.instru, 'f', -1, 64)},
			"target_key":              {strconv.Itoa(mid.key)},
			"target_loudness":         {strconv.FormatFloat(mid.loudness, 'f', -1, 64)},
			"target_tempo":            {strconv.FormatFloat(mid.tempo, 'f', -1, 64)},
			"target_valence":          {strconv.FormatFloat(mid.valence, 'f', -1, 64)},
		}.Encode(),
	)

	fmt.Printf("Da URL is: %s", url)

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return recommendationObject{}, errors.New("cannot generate request")
	}

	request.Header.Set(authorization, "Bearer "+accessToken)
	request.Header.Add("Content-Type", "application/json")

	result, err := client.Do(request)
	if err != nil {
		return recommendationObject{}, errors.New("request error")
	}
	responseBody, err := ReadAndCloseBody(result)
	if err != nil {
		return recommendationObject{}, errors.New("could not read response")
	}
	//fmt.Printf("Got Recos: %s", string(responseBody))

	var recos recommendationObject
	err = json.Unmarshal(responseBody, &recos)
	if err != nil {
		//logger.Log("err", err, "msg", fmt.Sprintf("user auth JSON error. The Response String is: %s", string(authResponseBody)))
		return recommendationObject{}, errors.New("decoding error")
	}

	return recos, nil
}

func createNewPlaylist(playlistName string) (string, error) {
	client := &http.Client{}

	u := fmt.Sprintf("https://api.spotify.com/v1/users/%s/playlists", myUserID)

	data := map[string]interface{}{
		"name":        playlistName,
		"description": "euhTest",
		"public":      false,
	}

	//fmt.Printf("url is: %s", u)
	d, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	request, err := http.NewRequest("POST", u, strings.NewReader(string(d)))
	if err != nil {
		return "", errors.New("cannot generate request")
	}

	request.Header.Set(authorization, "Bearer "+accessToken)
	request.Header.Add("Content-Type", "application/json")

	result, err := client.Do(request)
	if err != nil {
		return "", errors.New("request error")
	}
	responseBody, err := ReadAndCloseBody(result)
	if err != nil {
		return "", errors.New("could not read response")
	}
	//fmt.Printf("\nAdding a new playlist: %s", string(responseBody))

	var pl playlistObject
	err = json.Unmarshal(responseBody, &pl)
	if err != nil {
		//logger.Log("err", err, "msg", fmt.Sprintf("user auth JSON error. The Response String is: %s", string(authResponseBody)))
		return "", errors.New("decoding error")
	}
	if pl.ID != "" {
		return pl.ID, nil
	} else {
		return "", errors.New("No playlist created")
	}
}

func addTracksToPlaylist(playlistName string, songs []string) error {
	client := &http.Client{}

	u := fmt.Sprintf("https://api.spotify.com/v1/playlists/%s/tracks?%s", playlistName,
		url.Values{
			"uris": {strings.Join(songs, ",")},
		}.Encode())

	fmt.Printf("adding url is: %s", u)
	request, err := http.NewRequest("POST", u, nil)
	if err != nil {
		return errors.New("cannot generate request")
	}

	request.Header.Set(authorization, "Bearer "+accessToken)
	request.Header.Add("Content-Type", "application/json")

	result, err := client.Do(request)
	if err != nil {
		return errors.New("request error")
	}
	responseBody, err := ReadAndCloseBody(result)
	if err != nil {
		return errors.New("could not read response")
	}
	fmt.Printf("\nAdding to playlist : %s", string(responseBody))

	return nil
}

func getMySavedTracks(client *http.Client) (pageObject, error) {

	//https: //api.spotify.com/v1/audio-features/{id}
	//rqU := fmt.Sprintf("%s/%s", audioURL, songID)

	//https://api.spotify.com/v1/me/tracks?offset=0&limit=20
	//https://api.spotify.com/v1/me/tracks

	req := "https://api.spotify.com/v1/me/tracks?offset=57&limit=50"
	authRequest, err := http.NewRequest("GET", req, nil)
	if err != nil {
		return pageObject{}, errors.New("cannot generate request")
	}

	authRequest.Header.Set(authorization, "Bearer "+accessToken)
	authRequest.Header.Add("Content-Type", "application/json")

	authResult, err := client.Do(authRequest)
	if err != nil {
		return pageObject{}, errors.New("request error")
	}

	authResponseBody, err := ReadAndCloseBody(authResult)
	if err != nil {
		return pageObject{}, errors.New("could not read response")
	}
	//fmt.Printf("Got: %s", string(authResponseBody))

	var songF pageObject
	err = json.Unmarshal(authResponseBody, &songF)
	if err != nil {
		//logger.Log("err", err, "msg", fmt.Sprintf("user auth JSON error. The Response String is: %s", string(authResponseBody)))
		return pageObject{}, errors.New("decoding error")
	}

	return songF, nil
}

func ReadAndCloseBody(r *http.Response) ([]byte, error) {
	defer r.Body.Close()

	buf := bytes.NewBuffer(make([]byte, 0, r.ContentLength+bytes.MinRead))
	if _, err := buf.ReadFrom(r.Body); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func calculateAverages(sgs []spotySongFeatures) songAverages {
	deno := len(sgs)
	//mode is odality of track >> Major is represented by 1 and minor is 0.
	var d, e, loud, speech, acou, instru, live, valence, tempo float64
	var k int

	for _, c := range sgs {
		d += c.Danceability
		e += c.Energy
		k += c.Key
		live += c.Liveness
		loud += c.Loudness
		acou += c.Acousticness
		instru += c.Instrumentalness
		speech += c.Speechiness
		valence += c.Valence
		tempo += c.Tempo
	}

	return songAverages{
		danceability: d / float64(deno),
		energy:       e / float64(deno),
		key:          k / deno,
		liveness:     live / float64(deno),
		loudness:     loud / float64(deno),
		acousticness: acou / float64(deno),
		instru:       instru / float64(deno),
		speechiness:  speech / float64(deno),
		valence:      valence / float64(deno),
		tempo:        tempo / float64(deno),
	}
}

func calculateRanges(val float64) (lower, upper float64) {
	return val * float64(0.90), val * float64(1.10)
}

func isInRange(val, lower, upper float64) bool {
	return (val >= lower && val <= upper)
}
