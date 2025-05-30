package testdata

type TestDataService struct {
	rng              *Rng
	ServiceVersionID uint
	Editors          []uint
	Admins           []uint
}

func (s *TestDataService) GetRandomEditor() uint {
	return s.Editors[s.rng.Intn(len(s.Editors))]
}

func (s *TestDataService) GetRandomAdmin() uint {
	return s.Admins[s.rng.Intn(len(s.Admins))]
}

type Registry struct {
	services                     map[uint]*TestDataService
	serviceVersionIDs            []uint
	unpublishedServiceVersionIDs []uint
	serviceTypeIds               []uint
	users                        []uint
	rng                          *Rng
}

func NewRegistry(rng *Rng) *Registry {
	return &Registry{
		services:                     make(map[uint]*TestDataService),
		serviceVersionIDs:            make([]uint, 0),
		unpublishedServiceVersionIDs: make([]uint, 0),
		serviceTypeIds:               make([]uint, 0),
		users:                        make([]uint, 0),
		rng:                          rng,
	}
}

func (r *Registry) GetRandomUser() uint {
	return r.users[r.rng.Intn(len(r.users))]
}

func (r *Registry) GetRandomService() *TestDataService {
	return r.services[r.serviceVersionIDs[r.rng.Intn(len(r.serviceVersionIDs))]]
}

func (r *Registry) GetRandomUnpublishedService() *TestDataService {
	if len(r.unpublishedServiceVersionIDs) == 0 {
		return nil
	}

	return r.services[r.unpublishedServiceVersionIDs[r.rng.Intn(len(r.unpublishedServiceVersionIDs))]]
}

func (r *Registry) GetRandomServiceType() uint {
	return r.serviceTypeIds[r.rng.Intn(len(r.serviceTypeIds))]
}

func (r *Registry) RegisterUser(id uint) {
	r.users = append(r.users, id)
}

func (r *Registry) RegisterService(id uint, editors []uint, admins []uint) {
	r.services[id] = &TestDataService{
		ServiceVersionID: id,
		Editors:          editors,
		Admins:           admins,
		rng:              r.rng,
	}
	r.serviceVersionIDs = append(r.serviceVersionIDs, id)
	r.unpublishedServiceVersionIDs = append(r.unpublishedServiceVersionIDs, id)
}

func (r *Registry) RegisterServiceType(id uint) {
	r.serviceTypeIds = append(r.serviceTypeIds, id)
}
