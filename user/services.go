package user

type Service struct {
	repo Repo
}

func (s Service) Register(u *User) {

}

func NewService(r Repo) *Service {
	return &Service{repo: r}
}
