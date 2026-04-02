package user

// Service provides business logic for user operations.
type Service struct {
	repo Repository
}

// NewService creates a new user service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// GetAllUsers returns all users.
func (s *Service) GetAllUsers() ([]User, error) {
	return s.repo.FindAll()
}

// GetUserByID returns a user by their ID.
func (s *Service) GetUserByID(id uint) (*User, error) {
	return s.repo.FindByID(id)
}

// GetUserByEmail returns a user by their email.
func (s *Service) GetUserByEmail(email string) (*User, error) {
	return s.repo.FindByEmail(email)
}

// GetUserByGoogleID returns a user by their Google OAuth ID.
func (s *Service) GetUserByGoogleID(googleID string) (*User, error) {
	return s.repo.FindByGoogleID(googleID)
}

// GetUserBySpotifyID returns a user by their Spotify OAuth ID.
func (s *Service) GetUserBySpotifyID(spotifyID string) (*User, error) {
	return s.repo.FindBySpotifyID(spotifyID)
}

// CreateUser creates a new user in the database.
func (s *Service) CreateUser(user *User) error {
	return s.repo.Create(user)
}

// UpdateUser saves changes to an existing user.
func (s *Service) UpdateUser(user *User) error {
	return s.repo.Update(user)
}
