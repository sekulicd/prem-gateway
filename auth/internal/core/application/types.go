package application

type CreateApiKeyReq struct {
	IsRootKey        bool
	Services         []string
	RequestsPerRange int
	RangeInSeconds   int
}
