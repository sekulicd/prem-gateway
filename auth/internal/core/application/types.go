package application

type CreateApiKeyReq struct {
	Service          string
	RequestsPerRange int
	RangeInSeconds   int
}
