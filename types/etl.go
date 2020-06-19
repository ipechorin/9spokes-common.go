package types

//ETLMessage is an AMQP message structure referencing a connection document and containing the produced message that is later extracted, transformed, and persisted
type ETLMessage struct {
	Osp          string `json:"osp"`
	Datasource   string `json:"datasource"`
	ConnectionID string `json:"connectionId"`
	Company      string `json:"company,omitempty"`
	Index        string `json:"index,omitempty"`
	Period       string `json:"period,omitempty"`
	Storage      string `json:"storage,omitempty"`
}
