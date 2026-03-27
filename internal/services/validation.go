package services

type ValidationService interface {
	ValidateCURP(curp string) (bool, error)
	ValidateRFC(rfc string) (bool, error)
}

type validationService struct{}

func NewValidationService() ValidationService {
	return &validationService{}
}

func (s *validationService) ValidateCURP(curp string) (bool, error) {
	return len(curp) == 18, nil
}

func (s *validationService) ValidateRFC(rfc string) (bool, error) {
	return len(rfc) >= 12 && len(rfc) <= 13, nil
}