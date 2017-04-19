package edgegrid

type Resource struct {
	Complete chan bool
}

func (resource *Resource) Init() {
	resource.Complete = make(chan bool, 1)
}

func (resource *Resource) PostUnmarshalJSON() error {
	resource.Init()
	resource.Complete <- true
	return nil
}
