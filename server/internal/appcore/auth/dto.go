package appcore

type RegisterInput struct {
	Email       string
	Password    string
	DisplayName string
}

type LoginInput struct {
	Email    string
	Password string
}

type TokenOutput struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
}
