package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

type Payload struct {
	Image string `json:"image"`
}

type VisionReq struct {
	Requests []Request `json:"requests"`
}

type Request struct {
	Image    Image     `json:"image"`
	Features []Feature `json:"features"`
}

type Image struct {
	Content string `json:"content"`
}

type Feature struct {
	Type string `json:"type"`
}

type VisionResponses struct {
	Responses []Response `json:"responses"`
}

type Response struct {
	SafeSearchAnnotation SafeSearchAnnotation `json:"safeSearchAnnotation"`
	LabelAnnotations     []LabelAnnotation    `json:"labelAnnotations"`
}

type SafeSearchAnnotation struct {
	Violence string `json:"violence"`
}

type LabelAnnotation struct {
	Description string `json:"description"`
}

type FbSendMessageReq struct {
	MessagingType string    `json:"messaging_type"`
	Recipient     Recipient `json:"recipient"`
	Message       Message   `json:"message"`
}

type Recipient struct {
	Id string `json:"id"`
}

type Message struct {
	Text string `json:"text"`
}

func main() {
	address := fmt.Sprintf("%s:%s", "", "8080")

	router := mux.NewRouter()
	router.HandleFunc("/analyze-image", analyzeImage).Methods("POST")

	log.Printf("server running at %s", address)
	log.Fatal(http.ListenAndServe(address, router))
}

func analyzeImage(responseWriter http.ResponseWriter, request *http.Request) {
	var payload Payload
	decoder := json.NewDecoder(request.Body)

	if err := decoder.Decode(&payload); err != nil {
		createErrorResponse(responseWriter, http.StatusBadRequest, "Invalid request payload")
		return
	}

	defer request.Body.Close()

	analyze(payload.Image)

	createJsonResponse(responseWriter, http.StatusOK, payload)
}

func createErrorResponse(responseWriter http.ResponseWriter, code int, message string) {
	createJsonResponse(responseWriter, code, map[string]string{"error": message})
}

func createJsonResponse(responseWriter http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(code)
	responseWriter.Write(response)
}

func analyze(image string) {
	broadcastMessageUrl := fmt.Sprintf("https://vision.googleapis.com/v1/images:annotate?key=%s", os.Getenv("VISION_API_KEY"))

	img := Image{Content: image}

	var features []Feature
	features = append(features, Feature{Type: "WEB_DETECTION"})
	features = append(features, Feature{Type: "SAFE_SEARCH_DETECTION"})
	features = append(features, Feature{Type: "LABEL_DETECTION"})

	var requests []Request
	requests = append(requests, Request{Image: img, Features: features})

	req := VisionReq{
		Requests: requests,
	}

	payload, err := json.Marshal(req)
	res, err := http.Post(broadcastMessageUrl, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		log.Fatal(err)
	}

	defer res.Body.Close()

	responses := VisionResponses{}
	err = json.NewDecoder(res.Body).Decode(&responses)
	if err != nil {
		panic(err)
	}

	var strBuilder strings.Builder

	for _, response := range responses.Responses {
		log.Println("violence : ", response.SafeSearchAnnotation.Violence)
		log.Println("label : ", response.LabelAnnotations)

		strBuilder.WriteString("Violence : " + response.SafeSearchAnnotation.Violence + "\n\n")
		strBuilder.WriteString("Violence related objects detected: \n")

		if strings.Contains(fmt.Sprint(response.LabelAnnotations), "Knife") {
			strBuilder.WriteString("knife\n")
		}
	}

	message := strBuilder.String()

	var psIds []string
	psIds = append(psIds, "3164926763533661")
	psIds = append(psIds, "2261957987199163")
	psIds = append(psIds, "2588420581185410")
	psIds = append(psIds, "2183290295081563")
	psIds = append(psIds, "2172166666192906")

	for _, psId := range psIds {
		go sendToFb(message, psId)
	}

}

func sendToFb(message string, psId string) {
	log.Println(message)

	req := FbSendMessageReq{
		Message:       Message{Text: message},
		MessagingType: "RESPONSE",
		Recipient:     Recipient{Id: psId},
	}

	url := fmt.Sprintf("https://graph.facebook.com/v3.2/me/messages?access_token=%s", os.Getenv("FB_PAGE_ACCESS_TOKEN"))

	payload, err := json.Marshal(req)
	res, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		log.Fatal(err)
	}

	defer res.Body.Close()

	responseBytes, err := ioutil.ReadAll(res.Body)
	log.Printf("received response: %s, status code: %d", string(responseBytes), res.StatusCode)
}
