package web

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"regexp"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type WebServer struct {
	Tracer trace.Tracer
}

type ZipCodeInput struct {
	CEP string `json:"cep"`
}

type TemperatureResponse struct {
	City  string  `json:"city"`
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

func NewWebServer() *WebServer {
	return &WebServer{
		Tracer: otel.Tracer("input-service"),
	}
}

func (w *WebServer) Serve() {
	http.HandleFunc("/", w.handleZipCode)
	http.ListenAndServe(":8080", nil)
}

func (w *WebServer) handleZipCode(rw http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		rw.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	ctx, span := w.Tracer.Start(req.Context(), "handle-zipcode")
	defer span.End()

	var input ZipCodeInput
	if err := json.NewDecoder(req.Body).Decode(&input); err != nil {
		rw.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(rw).Encode(map[string]string{"error": "invalid zipcode"})
		return
	}

	if !isValidZipCode(input.CEP) {
		rw.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(rw).Encode(map[string]string{"error": "invalid zipcode"})
		return
	}

	_, childSpan := w.Tracer.Start(ctx, "call-temperature-service")
	defer childSpan.End()

	jsonData, _ := json.Marshal(input)
	resp, err := http.Post("http://temperature-service:8081/temperature", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(rw).Encode(map[string]string{"error": "failed to call temperature service"})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		rw.WriteHeader(resp.StatusCode)
		rw.Write(body)
		return
	}

	var tempResponse TemperatureResponse
	if err := json.Unmarshal(body, &tempResponse); err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(rw).Encode(map[string]string{"error": "failed to parse temperature service response"})
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(tempResponse)
}

func isValidZipCode(cep string) bool {
	match, _ := regexp.MatchString(`^\d{8}$`, cep)
	return match
}
