package entities

type Response struct {
    Status       int `json:"status"`
	Error 		 bool `json:"error"`
    Message   	 string `json:"message"`
    Data         any `json:"data"`
}

type ResponseWithPagination struct {
    Status       int `json:"status"`
	Error 		 bool `json:"error"`
    Message   	 string `json:"message"`
    Total        any `json:"total"`   
    PerPage      any `json:"per_page"`
    PrevPage     any `json:"prev_page"`
    NextPage     any `json:"next_page"`
    CurrentPage  any `json:"current_page"`
    NextUrl      any `json:"next_url"`
    PrevUrl      any `json:"prev_url"`
    Data         any `json:"data"`
}