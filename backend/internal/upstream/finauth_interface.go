package upstream

// FinAuthInterface FinAuth 客户端接口
type FinAuthInterface interface {
	GetToken(req *GetTokenRequest) (*GetTokenResponse, error)
	GetResult(req *GetResultRequest) (*GetResultResponse, error)
	GenerateSign() string
	VerifySign(jsonData, receivedSign string) bool
}

// 确保实现符合接口
var _ FinAuthInterface = (*FinAuthClient)(nil)
