package auth

import (
	"net/http"
	"strings"
)

type rules struct {
	resources map[string]*Resource
}

// gocsv 无法下载使用
// func NewRules(params ...interface{}) (Rules, error) {
// 	r := &rules{
// 		resources: make(map[string]*Resource),
// 	}
// 	resFileName := "./config/res.csv"
// 	if len(params) > 0 {
// 		rf, ok := params[0].(string)
// 		if ok {
// 			resFileName = rf
// 		}
// 	}

// 	resFile, err := os.OpenFile(resFileName, os.O_RDWR|os.O_CREATE, os.ModePerm)

// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resFile.Close()
// 	resources := make([]*Resource, 0)
// 	if err := gocsv.Unmarshal(resFile, &resources); err != nil {
// 		return nil, err
// 	}
// 	for _, res := range resources {
// 		r.resources[res.ID] = res
// 	}

// 	return r, nil
// }

func (r *rules) GetRes(id string) *Resource {
	return r.resources[id]
}

func (r *rules) Resources(domains ...string) ([]*Resource, error) {
	result := make([]*Resource, 0)
	for _, res := range r.resources {
		if len(domains) == 0 {
			result = append(result, res)
		} else {
			for _, d := range domains {
				dd := strings.Trim(res.Domain, " ")
				if strings.EqualFold(d, dd) {
					result = append(result, res)
				}
			}
		}
	}
	return result, nil

}

func (r *rules) VerifyRequest(req *http.Request) (bool, error) {
	if len(r.resources) > 0 {
		user, ok := AuthUser(req.Context())
		if ok {
			for _, res := range r.resources {
				if IsPatternMatch(strings.ToLower(req.RequestURI),
					strings.ToLower(res.EndPoint)) {
					return r.Verify(user, res)
				}
			}
		}
	}
	return false, nil
}

func (r *rules) Verify(user *Principal, res *Resource) (bool, error) {
	if len(user.Scopes) > 0 {
		if user.Scopes[0] == "*" {
			return true, nil
		}

		for _, sc := range user.Scopes {
			if strings.EqualFold(sc, res.ID) {
				return true, nil
			}
		}
	}
	return false, nil
}
