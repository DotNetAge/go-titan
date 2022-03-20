package middlewares

import (
	"context"
	"errors"
	"go-titan/auth"
	"go-titan/gateway"
	"net/http"
	"sort"
	"strings"

	"github.com/casbin/casbin/v2"
)

var (
	ErrBadAuthorizedToken      = errors.New("非法的访问令牌")
	ErrAuthorizedTokenNotFound = errors.New("未提供访问令牌")
)

const (
	HttpJSONContent     = "application/json"
	HttpFormContent     = "application/x-www-form-urlencoded"
	HttpFormDataContent = "application/form-data"
)

// AllowClients 用于过滤不包含在指定范围内的所有请求
// 如果成功过滤可通过XClientKey读取当访问的ClientID
func AllowClients(clients []string, ignores ...string) gateway.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if len(ignores) > 0 {
				for _, ipath := range ignores {
					if strings.HasPrefix(req.URL.Path, ipath) {
						next.ServeHTTP(w, req)
						return
					}
				}
			}
			// 从Header获取XClient
			val := req.Header.Get("X-Client")
			ckey := "client_id"
			// 从URL获取client_id
			if len(val) == 0 {
				val = req.URL.Query().Get(ckey)

				if len(val) == 0 {
					if req.Method == "POST" || req.Method == "PUT" {
						contentType := req.Header.Get("Content-Type")
						if strings.Contains(contentType, HttpFormContent) {
							// 从FORM中获取client_id
							val = req.FormValue(ckey)
						}

						if strings.Contains(contentType, HttpJSONContent) {
							mapResult, err := gateway.ReadDataFromBody(w, req)
							if err != nil {
								http.Error(w, err.Error(), http.StatusBadRequest)
								return
							}
							if mapResult[ckey] != nil {
								val = mapResult[ckey].(string)
							}
						}
					}
				}
			}

			// 如果不设置客户端检查就直接转入下个处理器
			if len(clients) == 0 {
				if len(val) == 0 {
					next.ServeHTTP(w, req)
					return
				} else {
					ctx := auth.ContextWithClient(req.Context(), val)
					next.ServeHTTP(w, req.WithContext(ctx))
					return
				}
			}

			if len(val) == 0 || !contains(clients, val) {
				http.Error(w, "未授权的客户端", http.StatusUnauthorized)
				return
			}
			ctx := auth.ContextWithClient(req.Context(), val)
			next.ServeHTTP(w, req.WithContext(ctx))
		})
	}
}

// TokenInspector 用户访问令牌读取中间件，从 Authorization 请求头中获取当前用户身份
// 如果成功读取 auth.IsAuth(ctx) 将会返回True
func TokenInspector(tokens auth.Tokens) gateway.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			val, err := authHeader(req, "bearer")
			if err == nil && val != "" {
				user, err := tokens.Inspect(val)
				if err == nil && user != nil {
					ctx := context.WithValue(req.Context(), auth.CurrentUserKey{}, user)
					next.ServeHTTP(w, req.WithContext(ctx))
					return
				}
			}

			next.ServeHTTP(w, req)

		})
	}
}

// Allows 过滤白名单外的全部请求
func Allows(ips ...string) gateway.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if !contains(ips, req.RemoteAddr) {
				http.Error(w, "未授权的IP", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, req)

		})
	}
}

// Blocks 拦截黑名单内的全部IP地址发出的请求
func Blocks(ips ...string) gateway.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if contains(ips, req.RemoteAddr) {
				http.Error(w, "IP已被封禁", http.StatusBadGateway)
				return
			}
			next.ServeHTTP(w, req)
		})
	}
}

// RolesInspector 基于Casbin进行RBAC角色验证
func RolesInspector(enforcer *casbin.Enforcer) gateway.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

			usr, ok := auth.AuthUser(req.Context())
			sub := "anonymous" // anonymous, administrators, members 为系统保留角色

			if ok { // 对已授权用户进行角色检查
				sub = "members" // 首先对已认证用户进行检查  usr.Name
			}

			domain := "admin.titan.com" // req.Host
			act := strings.ToLower(req.Method)
			obj := strings.ToLower(req.URL.Path)
			allows, err := enforcer.Enforce(sub, domain, obj, act)

			if err != nil {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
				return
			}

			if allows != true {

				if !ok {
					http.Error(w, "非法的用户身份", http.StatusUnauthorized)
					return
				}

				// 更严格的查询，如果没有通过认证用户则对具体用户名进行验证
				allows, err = enforcer.Enforce(usr.Name, domain, obj, act)

				if err != nil {
					http.Error(w, err.Error(), http.StatusServiceUnavailable)
					return
				}

			}

			if allows {
				next.ServeHTTP(w, req)
				return

			} else {
				// Access Deny
				http.Error(w, "当前用户角色对资源不具有访问权", http.StatusForbidden)
			}
		})
	}
}

// ScopeInspector Scopes 检查
func ScopeInspector(rules auth.Rules) gateway.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ok, err := rules.VerifyRequest(req)
			if err == nil {
				if !ok {
					http.Error(w, "Access Denied", http.StatusForbidden)
					return
				}
			}
			next.ServeHTTP(w, req)
		})
	}
}

func contains(s []string, searchterm string) bool {
	if len(s) == 0 || s == nil {
		return false
	}
	i := sort.SearchStrings(s, searchterm)
	return i < len(s) && s[i] == searchterm
}

func authHeader(req *http.Request, expectedScheme string) (string, error) {
	val := req.Header.Get("Authorization")
	if val == "" {
		return "", ErrAuthorizedTokenNotFound
	}

	splits := strings.SplitN(val, " ", 2)
	if len(splits) < 2 {
		return "", ErrBadAuthorizedToken
	}

	if !strings.EqualFold(splits[0], expectedScheme) {
		return "", ErrBadAuthorizedToken
	}

	return splits[1], nil
}
