package api

import (
	"encoding/json"
	"errors"
	"github.com/voodooEntity/archivist"
	"github.com/voodooEntity/gits/src/transport"
	"github.com/voodooEntity/gitsapi"
	"github.com/voodooEntity/go-cyberbrain/src/system/mapper"
	"github.com/voodooEntity/go-cyberbrain/src/system/registry"
	"github.com/voodooEntity/go-cyberbrain/src/system/scheduler"
	"io/ioutil"
	"net/http"
)

func Extend() {

	archivist.Info("> Extending gits-HTTP API")

	// Route: /v1/mapJson
	gitsapi.ServeMux.HandleFunc("/v1/learn", func(w http.ResponseWriter, r *http.Request) {
		if "OPTIONS" == r.Method {
			respond("", 200, w)
			return
		}

		// check http method
		if "POST" != r.Method {
			http.Error(w, "Invalid http method for this path", 422)
			return
		}

		// retrieve data from request
		body, err := getRequestBody(r)
		if nil != err {
			archivist.Error("Could not read http request body", err.Error())
			http.Error(w, "Malformed or no body. ", 422)
			return
		}

		// unpack the json
		var transportData transport.TransportEntity
		if err := json.Unmarshal(body, &transportData); err != nil {
			archivist.Error("Invalid json query object", err.Error())
			http.Error(w, "Invalid json query object "+err.Error(), 422)
			return
		}

		// lets pass the body to our mapper
		// that will recursive map the entities
		mappedData := mapper.MapTransportDataWithContext(transportData, "Data")
		scheduler.Run(mappedData, registry.Data)
		if nil != err {
			http.Error(w, err.Error(), 422)
			return
		}

		respondOk(transport.Transport{
			Entities: []transport.TransportEntity{mappedData},
		}, w)
	})
}

func getOptionalUrlParams(optionalUrlParams map[string]string, urlParams map[string]string, r *http.Request) map[string]string {
	tmpParams := r.URL.Query()
	for paramName := range optionalUrlParams {
		val, ok := tmpParams[paramName]
		if ok {
			urlParams[paramName] = val[0]
		}
	}
	return urlParams
}

func getRequiredUrlParams(requiredUrlParams map[string]string, r *http.Request) (map[string]string, error) {
	urlParams := r.URL.Query()
	for paramName := range requiredUrlParams {
		val, ok := urlParams[paramName]
		if !ok {
			return nil, errors.New("Missing required url param")
		}
		requiredUrlParams[paramName] = val[0]
	}
	return requiredUrlParams, nil
}

func respond(message string, responseCode int, w http.ResponseWriter) {

	w.Header().Add("Access-Control-Allow-Headers", "*")
	w.Header().Add("Access-Control-Allow-Origin", "*")

	w.WriteHeader(responseCode)
	messageBytes := []byte(message)

	_, err := w.Write(messageBytes)
	if nil != err {
		archivist.Error("Could not write http response body ", err, message)
	}
}

func respondOk(data transport.Transport, w http.ResponseWriter) {
	// than we gonne json encode it
	// build the json
	responseData, err := json.Marshal(data)
	if nil != err {
		http.Error(w, "Error building response data json", 500)
		return
	}

	// finally we gonne send our response
	w.Header().Add("Access-Control-Allow-Headers", "*")
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.WriteHeader(200)
	_, err = w.Write(responseData)
	if nil != err {
		archivist.Error("Could not write http response body ", err, data)
	}
}

func getRequestBody(r *http.Request) ([]byte, error) {
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return nil, err
	}
	return body, nil
}
