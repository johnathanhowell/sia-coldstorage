package main

import (
	"html/template"
	"log"
	"net"
	"net/http"

	"github.com/skratchdot/open-golang/open"

	"github.com/NebulousLabs/Sia/crypto"
	"github.com/NebulousLabs/Sia/modules"
	"github.com/NebulousLabs/Sia/types"
	"github.com/NebulousLabs/fastrand"
)

const outputTmpl = `
<html>
	<head>
		<title> Sia Cold Storage </title>
	</head>
	<body>
		<h3>Sia cold wallet successfully generated.</h3>
		<p> Please save the information below in a safe place. You can use the Seed to recover any money sent to any of the addresses, without the wallet being online. </p>
		<h4>Seed: </h4>
		<p>{{.Seed}}</p>
		<h4>Addresses: </h4>
		<ul>
		{{ range .Addresses }}
			<li>{{.}}</li>
		{{ end }}
	</body>
</html>
`

const nAddresses = 20

// getAddress returns an address generated from a seed at the index specified
// by `index`.
func getAddress(seed modules.Seed, index uint64) types.UnlockHash {
	_, pk := crypto.GenerateKeyPairDeterministic(crypto.HashAll(seed, index))
	return types.UnlockConditions{
		PublicKeys:         []types.SiaPublicKey{types.Ed25519PublicKey(pk)},
		SignaturesRequired: 1,
	}.UnlockHash()
}

func main() {
	// generate a seed and a few addresses from that seed
	var seed modules.Seed
	fastrand.Read(seed[:])
	var addresses []types.UnlockHash
	seedStr, err := modules.SeedToString(seed, "english")
	if err != nil {
		log.Fatal(err)
	}
	for i := uint64(0); i < nAddresses; i++ {
		addresses = append(addresses, getAddress(seed, i))
	}

	templateData := struct {
		Seed      string
		Addresses []types.UnlockHash
	}{
		Seed:      seedStr,
		Addresses: addresses,
	}
	t, err := template.New("output").Parse(outputTmpl)
	if err != nil {
		log.Fatal(err)
	}
	l, err := net.Listen("tcp", "localhost:8087")
	if err != nil {
		log.Fatal(err)
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Execute(w, templateData)
		l.Close()
	})
	go http.Serve(l, handler)

	open.Run("http://localhost:8087")
}
