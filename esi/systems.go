package esi

// func (e *Client) GetUniverseSystemsSystemID(id int) (Response, error) {

// 	url := url.URL{
// 		Scheme: "https",
// 		Host:   e.Host,
// 		Path:   fmt.Sprintf("/v4/universe/systems/%d/", id),
// 	}

// 	headers := make(map[string]string)

// 	request := Request{
// 		Method:  http.MethodGet,
// 		Path:    url,
// 		Headers: headers,
// 	}

// 	response, err := e.Request(request)
// 	if err != nil || response.Code >= 400 {
// 		return response, err
// 	}

// 	killmail := killboard.Killmail{}

// 	err = json.Unmarshal(response.Data.([]byte), &killmail)
// 	if err != nil {
// 		err = errors.Wrap(err, "unable to unmarshel response body")
// 		return response, err
// 	}

// 	response.Data = killmail

// 	return response, err
// }
