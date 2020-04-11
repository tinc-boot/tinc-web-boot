package network

//go:generate events-gen -p network -E Events -s -P -o events.go --ts ../web/ui/src/api/events.ts

//event:"Started"
//event:"Stopped"
type NetworkID struct {
	Name string `json:"name"`
}
