package web

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type WebServer struct {
	Tracer trace.Tracer
}

type ZipCodeInput struct {
	CEP string `json:"cep"`
}

type ViaCEPResponse struct {
	CEP         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	UF          string `json:"uf"`
}

type WeatherResponse struct {
	Current struct {
		TempC float64 `json:"temp_c"`
	} `json:"current"`
}

type TemperatureResponse struct {
	City  string  `json:"city"`
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

func NewWebServer() *WebServer {
	return &WebServer{
		Tracer: otel.Tracer("temperature-service"),
	}
}

func (w *WebServer) Serve() {
	http.HandleFunc("/temperature", w.handleTemperature)
	http.ListenAndServe(":8081", nil)
}

func (w *WebServer) handleTemperature(rw http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		rw.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	ctx, span := w.Tracer.Start(req.Context(), "handle-temperature")
	defer span.End()

	var input ZipCodeInput
	if err := json.NewDecoder(req.Body).Decode(&input); err != nil {
		rw.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(rw).Encode(map[string]string{"error": "invalid zipcode"})
		return
	}

	_, viaCEPSpan := w.Tracer.Start(ctx, "viacep-request")
	viaCEPResp, err := http.Get(fmt.Sprintf("http://viacep.com.br/ws/%s/json/", input.CEP))
	if err != nil {
		viaCEPSpan.End()
		rw.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(rw).Encode(map[string]string{"error": "failed to get address"})
		return
	}
	defer viaCEPResp.Body.Close()
	viaCEPSpan.End()

	var viacep ViaCEPResponse
	if err := json.NewDecoder(viaCEPResp.Body).Decode(&viacep); err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(rw).Encode(map[string]string{"error": "failed to parse address"})
		return
	}

	if viacep.Bairro == "" && viacep.UF == "" {
		rw.WriteHeader(http.StatusNotFound)
		json.NewEncoder(rw).Encode(map[string]string{"error": "can not find zipcode"})
		return
	}

	_, weatherSpan := w.Tracer.Start(ctx, "weather-request")
	weatherAPIKey := os.Getenv("WEATHER_API_KEY")
	encodedLocalidade := url.QueryEscape(viacep.Localidade)
	weatherURL := fmt.Sprintf("http://api.weatherapi.com/v1/current.json?key=%s&q=%s", weatherAPIKey, encodedLocalidade)
	fmt.Println("weatherURL", weatherURL)
	weatherResp, err := http.Get(weatherURL)
	if err != nil {
		weatherSpan.End()
		rw.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(rw).Encode(map[string]string{"error": "failed to get temperature"})
		return
	}
	defer weatherResp.Body.Close()
	weatherSpan.End()

	body, _ := io.ReadAll(weatherResp.Body)
	var weather WeatherResponse
	if err := json.Unmarshal(body, &weather); err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(rw).Encode(map[string]string{"error": "failed to parse temperature"})
		return
	}

	tempC := weather.Current.TempC
	tempF := tempC*1.8 + 32
	tempK := tempC + 273.0

	response := TemperatureResponse{
		City:  viacep.Localidade,
		TempC: tempC,
		TempF: tempF,
		TempK: tempK,
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(response)
}
