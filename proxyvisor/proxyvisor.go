package proxyvisor

func Provision() (string, error) {
	return "PIo0c6fYkQN0liVOPAVs7WA4WtP8NYChlDjRKPO75xpmT1Kj0wk1nPJFIwDStv", nil
	// resp, err := http.Post("http://localhost:3000/v1/device", "", nil)
	// defer resp.Body.Close()

	// if err != nil {
	// 	log.Println("Failed to query proxyvisor")
	// 	return "", err
	// }

	// b, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	log.Println("Failed to parse response")
	// 	return "", err
	// }
	// return string(b), nil
}
