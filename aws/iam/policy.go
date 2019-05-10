package iam

// normalizeInterfaceToStringArr normalizes a single string or []string to []string
func normalizeInterfaceToStringArr(inter interface{}) []string {
	var sArr []string

	switch inter.(type) {
	case string:
		sArr = []string{inter.(string)}
	case []interface{}:
		sInters := inter.([]interface{})

		for _, i := range sInters {
			s, ok := i.(string)
			if ok {
				sArr = append(sArr, s)
			}
		}
	}

	return sArr
}

type PolicyDocument struct {
	Version   string
	Statement []StatementEntry
}

type StatementEntry struct {
	Effect    string
	Action    interface{}
	Resource  string
	Principal PrincipalEntry
}

func (s StatementEntry) NormalizedAction() []string {
	return normalizeInterfaceToStringArr(s.Action)
}

type PrincipalEntry struct {
	AWS interface{}
}

// AWS normalizes the string or []string entries in AWS principals.
func (e PrincipalEntry) NormalizedAWS() []string {
	return normalizeInterfaceToStringArr(e.AWS)
}
