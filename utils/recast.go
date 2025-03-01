package utils

func Recast(a, b any) error {
	js, err := json.Marshal(a)
	if err != nil {
		return err
	}
	return json.Unmarshal(js, b)
}
