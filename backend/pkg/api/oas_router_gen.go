// Code generated by ogen, DO NOT EDIT.

package api

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/ogen-go/ogen/uri"
)

func (s *Server) cutPrefix(path string) (string, bool) {
	prefix := s.cfg.Prefix
	if prefix == "" {
		return path, true
	}
	if !strings.HasPrefix(path, prefix) {
		// Prefix doesn't match.
		return "", false
	}
	// Cut prefix from the path.
	return strings.TrimPrefix(path, prefix), true
}

// ServeHTTP serves http request as defined by OpenAPI v3 specification,
// calling handler that matches the path or returning not found error.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	elem := r.URL.Path
	elemIsEscaped := false
	if rawPath := r.URL.RawPath; rawPath != "" {
		if normalized, ok := uri.NormalizeEscapedPath(rawPath); ok {
			elem = normalized
			elemIsEscaped = strings.ContainsRune(elem, '%')
		}
	}

	elem, ok := s.cutPrefix(elem)
	if !ok || len(elem) == 0 {
		s.notFound(w, r)
		return
	}
	args := [2]string{}

	// Static code generated router with unwrapped path search.
	switch {
	default:
		if len(elem) == 0 {
			break
		}
		switch elem[0] {
		case '/': // Prefix: "/"
			origElem := elem
			if l := len("/"); len(elem) >= l && elem[0:l] == "/" {
				elem = elem[l:]
			} else {
				break
			}

			if len(elem) == 0 {
				break
			}
			switch elem[0] {
			case 'd': // Prefix: "datasource/"
				origElem := elem
				if l := len("datasource/"); len(elem) >= l && elem[0:l] == "datasource/" {
					elem = elem[l:]
				} else {
					break
				}

				if len(elem) == 0 {
					break
				}
				switch elem[0] {
				case 'e': // Prefix: "email"
					origElem := elem
					if l := len("email"); len(elem) >= l && elem[0:l] == "email" {
						elem = elem[l:]
					} else {
						break
					}

					if len(elem) == 0 {
						switch r.Method {
						case "GET":
							s.handleDatasourceEmailListRequest([0]string{}, elemIsEscaped, w, r)
						case "POST":
							s.handleDatasourceEmailCreateRequest([0]string{}, elemIsEscaped, w, r)
						default:
							s.notAllowed(w, r, "GET,POST")
						}

						return
					}
					switch elem[0] {
					case '/': // Prefix: "/"
						origElem := elem
						if l := len("/"); len(elem) >= l && elem[0:l] == "/" {
							elem = elem[l:]
						} else {
							break
						}

						// Param: "uuid"
						// Match until "/"
						idx := strings.IndexByte(elem, '/')
						if idx < 0 {
							idx = len(elem)
						}
						args[0] = elem[:idx]
						elem = elem[idx:]

						if len(elem) == 0 {
							switch r.Method {
							case "DELETE":
								s.handleDatasourceEmailDeleteRequest([1]string{
									args[0],
								}, elemIsEscaped, w, r)
							case "GET":
								s.handleDatasourceEmailGetRequest([1]string{
									args[0],
								}, elemIsEscaped, w, r)
							case "PUT":
								s.handleDatasourceEmailUpdateRequest([1]string{
									args[0],
								}, elemIsEscaped, w, r)
							default:
								s.notAllowed(w, r, "DELETE,GET,PUT")
							}

							return
						}
						switch elem[0] {
						case '/': // Prefix: "/run/pipeline"
							origElem := elem
							if l := len("/run/pipeline"); len(elem) >= l && elem[0:l] == "/run/pipeline" {
								elem = elem[l:]
							} else {
								break
							}

							if len(elem) == 0 {
								// Leaf node.
								switch r.Method {
								case "POST":
									s.handleDatasourceEmailRunPipelineRequest([1]string{
										args[0],
									}, elemIsEscaped, w, r)
								default:
									s.notAllowed(w, r, "POST")
								}

								return
							}

							elem = origElem
						}

						elem = origElem
					}

					elem = origElem
				}
				// Param: "uuid"
				// Match until "/"
				idx := strings.IndexByte(elem, '/')
				if idx < 0 {
					idx = len(elem)
				}
				args[0] = elem[:idx]
				elem = elem[idx:]

				if len(elem) == 0 {
					break
				}
				switch elem[0] {
				case '/': // Prefix: "/oauth2/client"
					origElem := elem
					if l := len("/oauth2/client"); len(elem) >= l && elem[0:l] == "/oauth2/client" {
						elem = elem[l:]
					} else {
						break
					}

					if len(elem) == 0 {
						// Leaf node.
						switch r.Method {
						case "PUT":
							s.handleDatasourceSetOAuth2ClientRequest([1]string{
								args[0],
							}, elemIsEscaped, w, r)
						default:
							s.notAllowed(w, r, "PUT")
						}

						return
					}

					elem = origElem
				}

				elem = origElem
			case 'o': // Prefix: "oauth2/"
				origElem := elem
				if l := len("oauth2/"); len(elem) >= l && elem[0:l] == "oauth2/" {
					elem = elem[l:]
				} else {
					break
				}

				if len(elem) == 0 {
					break
				}
				switch elem[0] {
				case 'c': // Prefix: "c"
					origElem := elem
					if l := len("c"); len(elem) >= l && elem[0:l] == "c" {
						elem = elem[l:]
					} else {
						break
					}

					if len(elem) == 0 {
						break
					}
					switch elem[0] {
					case 'a': // Prefix: "allback"
						origElem := elem
						if l := len("allback"); len(elem) >= l && elem[0:l] == "allback" {
							elem = elem[l:]
						} else {
							break
						}

						if len(elem) == 0 {
							// Leaf node.
							switch r.Method {
							case "GET":
								s.handleOAuth2ClientCallbackRequest([0]string{}, elemIsEscaped, w, r)
							default:
								s.notAllowed(w, r, "GET")
							}

							return
						}

						elem = origElem
					case 'l': // Prefix: "lient"
						origElem := elem
						if l := len("lient"); len(elem) >= l && elem[0:l] == "lient" {
							elem = elem[l:]
						} else {
							break
						}

						if len(elem) == 0 {
							switch r.Method {
							case "GET":
								s.handleOAuth2ClientListRequest([0]string{}, elemIsEscaped, w, r)
							case "POST":
								s.handleOAuth2ClientCreateRequest([0]string{}, elemIsEscaped, w, r)
							default:
								s.notAllowed(w, r, "GET,POST")
							}

							return
						}
						switch elem[0] {
						case '/': // Prefix: "/"
							origElem := elem
							if l := len("/"); len(elem) >= l && elem[0:l] == "/" {
								elem = elem[l:]
							} else {
								break
							}

							// Param: "id"
							// Match until "/"
							idx := strings.IndexByte(elem, '/')
							if idx < 0 {
								idx = len(elem)
							}
							args[0] = elem[:idx]
							elem = elem[idx:]

							if len(elem) == 0 {
								switch r.Method {
								case "DELETE":
									s.handleOAuth2ClientDeleteRequest([1]string{
										args[0],
									}, elemIsEscaped, w, r)
								case "GET":
									s.handleOAuth2ClientGetRequest([1]string{
										args[0],
									}, elemIsEscaped, w, r)
								case "PUT":
									s.handleOAuth2ClientUpdateRequest([1]string{
										args[0],
									}, elemIsEscaped, w, r)
								default:
									s.notAllowed(w, r, "DELETE,GET,PUT")
								}

								return
							}
							switch elem[0] {
							case '/': // Prefix: "/token"
								origElem := elem
								if l := len("/token"); len(elem) >= l && elem[0:l] == "/token" {
									elem = elem[l:]
								} else {
									break
								}

								if len(elem) == 0 {
									switch r.Method {
									case "GET":
										s.handleOAuth2ClientTokenListRequest([1]string{
											args[0],
										}, elemIsEscaped, w, r)
									default:
										s.notAllowed(w, r, "GET")
									}

									return
								}
								switch elem[0] {
								case '/': // Prefix: "/"
									origElem := elem
									if l := len("/"); len(elem) >= l && elem[0:l] == "/" {
										elem = elem[l:]
									} else {
										break
									}

									// Param: "uuid"
									// Leaf parameter
									args[1] = elem
									elem = ""

									if len(elem) == 0 {
										// Leaf node.
										switch r.Method {
										case "DELETE":
											s.handleOAuth2ClientTokenDeleteRequest([2]string{
												args[0],
												args[1],
											}, elemIsEscaped, w, r)
										default:
											s.notAllowed(w, r, "DELETE")
										}

										return
									}

									elem = origElem
								}

								elem = origElem
							}

							elem = origElem
						}

						elem = origElem
					}

					elem = origElem
				case 'l': // Prefix: "login"
					origElem := elem
					if l := len("login"); len(elem) >= l && elem[0:l] == "login" {
						elem = elem[l:]
					} else {
						break
					}

					if len(elem) == 0 {
						// Leaf node.
						switch r.Method {
						case "POST":
							s.handleOAuth2ClientLoginRequest([0]string{}, elemIsEscaped, w, r)
						default:
							s.notAllowed(w, r, "POST")
						}

						return
					}

					elem = origElem
				}

				elem = origElem
			case 'p': // Prefix: "pipeline"
				origElem := elem
				if l := len("pipeline"); len(elem) >= l && elem[0:l] == "pipeline" {
					elem = elem[l:]
				} else {
					break
				}

				if len(elem) == 0 {
					switch r.Method {
					case "GET":
						s.handlePipelineListRequest([0]string{}, elemIsEscaped, w, r)
					case "POST":
						s.handlePipelineCreateRequest([0]string{}, elemIsEscaped, w, r)
					default:
						s.notAllowed(w, r, "GET,POST")
					}

					return
				}
				switch elem[0] {
				case '/': // Prefix: "/"
					origElem := elem
					if l := len("/"); len(elem) >= l && elem[0:l] == "/" {
						elem = elem[l:]
					} else {
						break
					}

					if len(elem) == 0 {
						break
					}
					switch elem[0] {
					case 'e': // Prefix: "entry/types"
						origElem := elem
						if l := len("entry/types"); len(elem) >= l && elem[0:l] == "entry/types" {
							elem = elem[l:]
						} else {
							break
						}

						if len(elem) == 0 {
							// Leaf node.
							switch r.Method {
							case "GET":
								s.handlePipelineEntryTypeListRequest([0]string{}, elemIsEscaped, w, r)
							default:
								s.notAllowed(w, r, "GET")
							}

							return
						}

						elem = origElem
					}
					// Param: "uuid"
					// Match until "/"
					idx := strings.IndexByte(elem, '/')
					if idx < 0 {
						idx = len(elem)
					}
					args[0] = elem[:idx]
					elem = elem[idx:]

					if len(elem) == 0 {
						switch r.Method {
						case "DELETE":
							s.handlePipelineDeleteRequest([1]string{
								args[0],
							}, elemIsEscaped, w, r)
						case "GET":
							s.handlePipelineGetRequest([1]string{
								args[0],
							}, elemIsEscaped, w, r)
						case "PUT":
							s.handlePipelineUpdateRequest([1]string{
								args[0],
							}, elemIsEscaped, w, r)
						default:
							s.notAllowed(w, r, "DELETE,GET,PUT")
						}

						return
					}
					switch elem[0] {
					case '/': // Prefix: "/entry"
						origElem := elem
						if l := len("/entry"); len(elem) >= l && elem[0:l] == "/entry" {
							elem = elem[l:]
						} else {
							break
						}

						if len(elem) == 0 {
							switch r.Method {
							case "GET":
								s.handlePipelineEntryListRequest([1]string{
									args[0],
								}, elemIsEscaped, w, r)
							case "POST":
								s.handlePipelineEntryCreateRequest([1]string{
									args[0],
								}, elemIsEscaped, w, r)
							default:
								s.notAllowed(w, r, "GET,POST")
							}

							return
						}
						switch elem[0] {
						case '/': // Prefix: "/"
							origElem := elem
							if l := len("/"); len(elem) >= l && elem[0:l] == "/" {
								elem = elem[l:]
							} else {
								break
							}

							// Param: "entry_uuid"
							// Leaf parameter
							args[1] = elem
							elem = ""

							if len(elem) == 0 {
								// Leaf node.
								switch r.Method {
								case "DELETE":
									s.handlePipelineEntryDeleteRequest([2]string{
										args[0],
										args[1],
									}, elemIsEscaped, w, r)
								case "GET":
									s.handlePipelineEntryGetRequest([2]string{
										args[0],
										args[1],
									}, elemIsEscaped, w, r)
								case "PUT":
									s.handlePipelineEntryUpdateRequest([2]string{
										args[0],
										args[1],
									}, elemIsEscaped, w, r)
								default:
									s.notAllowed(w, r, "DELETE,GET,PUT")
								}

								return
							}

							elem = origElem
						}

						elem = origElem
					}

					elem = origElem
				}

				elem = origElem
			case 's': // Prefix: "storage"
				origElem := elem
				if l := len("storage"); len(elem) >= l && elem[0:l] == "storage" {
					elem = elem[l:]
				} else {
					break
				}

				if len(elem) == 0 {
					switch r.Method {
					case "GET":
						s.handleStorageListRequest([0]string{}, elemIsEscaped, w, r)
					default:
						s.notAllowed(w, r, "GET")
					}

					return
				}
				switch elem[0] {
				case '/': // Prefix: "/postgres"
					origElem := elem
					if l := len("/postgres"); len(elem) >= l && elem[0:l] == "/postgres" {
						elem = elem[l:]
					} else {
						break
					}

					if len(elem) == 0 {
						switch r.Method {
						case "POST":
							s.handleStoragePostgresCreateRequest([0]string{}, elemIsEscaped, w, r)
						default:
							s.notAllowed(w, r, "POST")
						}

						return
					}
					switch elem[0] {
					case '/': // Prefix: "/"
						origElem := elem
						if l := len("/"); len(elem) >= l && elem[0:l] == "/" {
							elem = elem[l:]
						} else {
							break
						}

						// Param: "uuid"
						// Leaf parameter
						args[0] = elem
						elem = ""

						if len(elem) == 0 {
							// Leaf node.
							switch r.Method {
							case "DELETE":
								s.handleStoragePostgresDeleteRequest([1]string{
									args[0],
								}, elemIsEscaped, w, r)
							case "GET":
								s.handleStoragePostgresGetRequest([1]string{
									args[0],
								}, elemIsEscaped, w, r)
							case "PUT":
								s.handleStoragePostgresUpdateRequest([1]string{
									args[0],
								}, elemIsEscaped, w, r)
							default:
								s.notAllowed(w, r, "DELETE,GET,PUT")
							}

							return
						}

						elem = origElem
					}

					elem = origElem
				}

				elem = origElem
			case 't': // Prefix: "tg"
				origElem := elem
				if l := len("tg"); len(elem) >= l && elem[0:l] == "tg" {
					elem = elem[l:]
				} else {
					break
				}

				if len(elem) == 0 {
					// Leaf node.
					switch r.Method {
					case "GET":
						s.handleTgSessionListRequest([0]string{}, elemIsEscaped, w, r)
					case "POST":
						s.handleTgSessionCreateRequest([0]string{}, elemIsEscaped, w, r)
					case "PUT":
						s.handleTgSessionVerifyRequest([0]string{}, elemIsEscaped, w, r)
					default:
						s.notAllowed(w, r, "GET,POST,PUT")
					}

					return
				}

				elem = origElem
			}

			elem = origElem
		}
	}
	s.notFound(w, r)
}

// Route is route object.
type Route struct {
	name        string
	summary     string
	operationID string
	pathPattern string
	count       int
	args        [2]string
}

// Name returns ogen operation name.
//
// It is guaranteed to be unique and not empty.
func (r Route) Name() string {
	return r.name
}

// Summary returns OpenAPI summary.
func (r Route) Summary() string {
	return r.summary
}

// OperationID returns OpenAPI operationId.
func (r Route) OperationID() string {
	return r.operationID
}

// PathPattern returns OpenAPI path.
func (r Route) PathPattern() string {
	return r.pathPattern
}

// Args returns parsed arguments.
func (r Route) Args() []string {
	return r.args[:r.count]
}

// FindRoute finds Route for given method and path.
//
// Note: this method does not unescape path or handle reserved characters in path properly. Use FindPath instead.
func (s *Server) FindRoute(method, path string) (Route, bool) {
	return s.FindPath(method, &url.URL{Path: path})
}

// FindPath finds Route for given method and URL.
func (s *Server) FindPath(method string, u *url.URL) (r Route, _ bool) {
	var (
		elem = u.Path
		args = r.args
	)
	if rawPath := u.RawPath; rawPath != "" {
		if normalized, ok := uri.NormalizeEscapedPath(rawPath); ok {
			elem = normalized
		}
		defer func() {
			for i, arg := range r.args[:r.count] {
				if unescaped, err := url.PathUnescape(arg); err == nil {
					r.args[i] = unescaped
				}
			}
		}()
	}

	elem, ok := s.cutPrefix(elem)
	if !ok {
		return r, false
	}

	// Static code generated router with unwrapped path search.
	switch {
	default:
		if len(elem) == 0 {
			break
		}
		switch elem[0] {
		case '/': // Prefix: "/"
			origElem := elem
			if l := len("/"); len(elem) >= l && elem[0:l] == "/" {
				elem = elem[l:]
			} else {
				break
			}

			if len(elem) == 0 {
				break
			}
			switch elem[0] {
			case 'd': // Prefix: "datasource/"
				origElem := elem
				if l := len("datasource/"); len(elem) >= l && elem[0:l] == "datasource/" {
					elem = elem[l:]
				} else {
					break
				}

				if len(elem) == 0 {
					break
				}
				switch elem[0] {
				case 'e': // Prefix: "email"
					origElem := elem
					if l := len("email"); len(elem) >= l && elem[0:l] == "email" {
						elem = elem[l:]
					} else {
						break
					}

					if len(elem) == 0 {
						switch method {
						case "GET":
							r.name = DatasourceEmailListOperation
							r.summary = ""
							r.operationID = "datasource-email-list"
							r.pathPattern = "/datasource/email"
							r.args = args
							r.count = 0
							return r, true
						case "POST":
							r.name = DatasourceEmailCreateOperation
							r.summary = ""
							r.operationID = "datasource-email-create"
							r.pathPattern = "/datasource/email"
							r.args = args
							r.count = 0
							return r, true
						default:
							return
						}
					}
					switch elem[0] {
					case '/': // Prefix: "/"
						origElem := elem
						if l := len("/"); len(elem) >= l && elem[0:l] == "/" {
							elem = elem[l:]
						} else {
							break
						}

						// Param: "uuid"
						// Match until "/"
						idx := strings.IndexByte(elem, '/')
						if idx < 0 {
							idx = len(elem)
						}
						args[0] = elem[:idx]
						elem = elem[idx:]

						if len(elem) == 0 {
							switch method {
							case "DELETE":
								r.name = DatasourceEmailDeleteOperation
								r.summary = ""
								r.operationID = "datasource-email-delete"
								r.pathPattern = "/datasource/email/{uuid}"
								r.args = args
								r.count = 1
								return r, true
							case "GET":
								r.name = DatasourceEmailGetOperation
								r.summary = ""
								r.operationID = "datasource-email-get"
								r.pathPattern = "/datasource/email/{uuid}"
								r.args = args
								r.count = 1
								return r, true
							case "PUT":
								r.name = DatasourceEmailUpdateOperation
								r.summary = ""
								r.operationID = "datasource-email-update"
								r.pathPattern = "/datasource/email/{uuid}"
								r.args = args
								r.count = 1
								return r, true
							default:
								return
							}
						}
						switch elem[0] {
						case '/': // Prefix: "/run/pipeline"
							origElem := elem
							if l := len("/run/pipeline"); len(elem) >= l && elem[0:l] == "/run/pipeline" {
								elem = elem[l:]
							} else {
								break
							}

							if len(elem) == 0 {
								// Leaf node.
								switch method {
								case "POST":
									r.name = DatasourceEmailRunPipelineOperation
									r.summary = ""
									r.operationID = "datasource-email-run-pipeline"
									r.pathPattern = "/datasource/email/{uuid}/run/pipeline"
									r.args = args
									r.count = 1
									return r, true
								default:
									return
								}
							}

							elem = origElem
						}

						elem = origElem
					}

					elem = origElem
				}
				// Param: "uuid"
				// Match until "/"
				idx := strings.IndexByte(elem, '/')
				if idx < 0 {
					idx = len(elem)
				}
				args[0] = elem[:idx]
				elem = elem[idx:]

				if len(elem) == 0 {
					break
				}
				switch elem[0] {
				case '/': // Prefix: "/oauth2/client"
					origElem := elem
					if l := len("/oauth2/client"); len(elem) >= l && elem[0:l] == "/oauth2/client" {
						elem = elem[l:]
					} else {
						break
					}

					if len(elem) == 0 {
						// Leaf node.
						switch method {
						case "PUT":
							r.name = DatasourceSetOAuth2ClientOperation
							r.summary = ""
							r.operationID = "datasource-set-oauth2-client"
							r.pathPattern = "/datasource/{uuid}/oauth2/client"
							r.args = args
							r.count = 1
							return r, true
						default:
							return
						}
					}

					elem = origElem
				}

				elem = origElem
			case 'o': // Prefix: "oauth2/"
				origElem := elem
				if l := len("oauth2/"); len(elem) >= l && elem[0:l] == "oauth2/" {
					elem = elem[l:]
				} else {
					break
				}

				if len(elem) == 0 {
					break
				}
				switch elem[0] {
				case 'c': // Prefix: "c"
					origElem := elem
					if l := len("c"); len(elem) >= l && elem[0:l] == "c" {
						elem = elem[l:]
					} else {
						break
					}

					if len(elem) == 0 {
						break
					}
					switch elem[0] {
					case 'a': // Prefix: "allback"
						origElem := elem
						if l := len("allback"); len(elem) >= l && elem[0:l] == "allback" {
							elem = elem[l:]
						} else {
							break
						}

						if len(elem) == 0 {
							// Leaf node.
							switch method {
							case "GET":
								r.name = OAuth2ClientCallbackOperation
								r.summary = ""
								r.operationID = "oauth2-client-callback"
								r.pathPattern = "/oauth2/callback"
								r.args = args
								r.count = 0
								return r, true
							default:
								return
							}
						}

						elem = origElem
					case 'l': // Prefix: "lient"
						origElem := elem
						if l := len("lient"); len(elem) >= l && elem[0:l] == "lient" {
							elem = elem[l:]
						} else {
							break
						}

						if len(elem) == 0 {
							switch method {
							case "GET":
								r.name = OAuth2ClientListOperation
								r.summary = ""
								r.operationID = "oauth2-client-list"
								r.pathPattern = "/oauth2/client"
								r.args = args
								r.count = 0
								return r, true
							case "POST":
								r.name = OAuth2ClientCreateOperation
								r.summary = ""
								r.operationID = "oauth2-client-create"
								r.pathPattern = "/oauth2/client"
								r.args = args
								r.count = 0
								return r, true
							default:
								return
							}
						}
						switch elem[0] {
						case '/': // Prefix: "/"
							origElem := elem
							if l := len("/"); len(elem) >= l && elem[0:l] == "/" {
								elem = elem[l:]
							} else {
								break
							}

							// Param: "id"
							// Match until "/"
							idx := strings.IndexByte(elem, '/')
							if idx < 0 {
								idx = len(elem)
							}
							args[0] = elem[:idx]
							elem = elem[idx:]

							if len(elem) == 0 {
								switch method {
								case "DELETE":
									r.name = OAuth2ClientDeleteOperation
									r.summary = ""
									r.operationID = "oauth2-client-delete"
									r.pathPattern = "/oauth2/client/{id}"
									r.args = args
									r.count = 1
									return r, true
								case "GET":
									r.name = OAuth2ClientGetOperation
									r.summary = ""
									r.operationID = "oauth2-client-get"
									r.pathPattern = "/oauth2/client/{id}"
									r.args = args
									r.count = 1
									return r, true
								case "PUT":
									r.name = OAuth2ClientUpdateOperation
									r.summary = ""
									r.operationID = "oauth2-client-update"
									r.pathPattern = "/oauth2/client/{id}"
									r.args = args
									r.count = 1
									return r, true
								default:
									return
								}
							}
							switch elem[0] {
							case '/': // Prefix: "/token"
								origElem := elem
								if l := len("/token"); len(elem) >= l && elem[0:l] == "/token" {
									elem = elem[l:]
								} else {
									break
								}

								if len(elem) == 0 {
									switch method {
									case "GET":
										r.name = OAuth2ClientTokenListOperation
										r.summary = ""
										r.operationID = "oauth2-client-token-list"
										r.pathPattern = "/oauth2/client/{datasource_uuid}/token"
										r.args = args
										r.count = 1
										return r, true
									default:
										return
									}
								}
								switch elem[0] {
								case '/': // Prefix: "/"
									origElem := elem
									if l := len("/"); len(elem) >= l && elem[0:l] == "/" {
										elem = elem[l:]
									} else {
										break
									}

									// Param: "uuid"
									// Leaf parameter
									args[1] = elem
									elem = ""

									if len(elem) == 0 {
										// Leaf node.
										switch method {
										case "DELETE":
											r.name = OAuth2ClientTokenDeleteOperation
											r.summary = ""
											r.operationID = "oauth2-client-token-delete"
											r.pathPattern = "/oauth2/client/{datasource_uuid}/token/{uuid}"
											r.args = args
											r.count = 2
											return r, true
										default:
											return
										}
									}

									elem = origElem
								}

								elem = origElem
							}

							elem = origElem
						}

						elem = origElem
					}

					elem = origElem
				case 'l': // Prefix: "login"
					origElem := elem
					if l := len("login"); len(elem) >= l && elem[0:l] == "login" {
						elem = elem[l:]
					} else {
						break
					}

					if len(elem) == 0 {
						// Leaf node.
						switch method {
						case "POST":
							r.name = OAuth2ClientLoginOperation
							r.summary = ""
							r.operationID = "oauth2-client-login"
							r.pathPattern = "/oauth2/login"
							r.args = args
							r.count = 0
							return r, true
						default:
							return
						}
					}

					elem = origElem
				}

				elem = origElem
			case 'p': // Prefix: "pipeline"
				origElem := elem
				if l := len("pipeline"); len(elem) >= l && elem[0:l] == "pipeline" {
					elem = elem[l:]
				} else {
					break
				}

				if len(elem) == 0 {
					switch method {
					case "GET":
						r.name = PipelineListOperation
						r.summary = ""
						r.operationID = "pipeline-list"
						r.pathPattern = "/pipeline"
						r.args = args
						r.count = 0
						return r, true
					case "POST":
						r.name = PipelineCreateOperation
						r.summary = ""
						r.operationID = "pipeline-create"
						r.pathPattern = "/pipeline"
						r.args = args
						r.count = 0
						return r, true
					default:
						return
					}
				}
				switch elem[0] {
				case '/': // Prefix: "/"
					origElem := elem
					if l := len("/"); len(elem) >= l && elem[0:l] == "/" {
						elem = elem[l:]
					} else {
						break
					}

					if len(elem) == 0 {
						break
					}
					switch elem[0] {
					case 'e': // Prefix: "entry/types"
						origElem := elem
						if l := len("entry/types"); len(elem) >= l && elem[0:l] == "entry/types" {
							elem = elem[l:]
						} else {
							break
						}

						if len(elem) == 0 {
							// Leaf node.
							switch method {
							case "GET":
								r.name = PipelineEntryTypeListOperation
								r.summary = ""
								r.operationID = "pipeline-entry-type-list"
								r.pathPattern = "/pipeline/entry/types"
								r.args = args
								r.count = 0
								return r, true
							default:
								return
							}
						}

						elem = origElem
					}
					// Param: "uuid"
					// Match until "/"
					idx := strings.IndexByte(elem, '/')
					if idx < 0 {
						idx = len(elem)
					}
					args[0] = elem[:idx]
					elem = elem[idx:]

					if len(elem) == 0 {
						switch method {
						case "DELETE":
							r.name = PipelineDeleteOperation
							r.summary = ""
							r.operationID = "pipeline-delete"
							r.pathPattern = "/pipeline/{uuid}"
							r.args = args
							r.count = 1
							return r, true
						case "GET":
							r.name = PipelineGetOperation
							r.summary = ""
							r.operationID = "pipeline-get"
							r.pathPattern = "/pipeline/{uuid}"
							r.args = args
							r.count = 1
							return r, true
						case "PUT":
							r.name = PipelineUpdateOperation
							r.summary = ""
							r.operationID = "pipeline-update"
							r.pathPattern = "/pipeline/{uuid}"
							r.args = args
							r.count = 1
							return r, true
						default:
							return
						}
					}
					switch elem[0] {
					case '/': // Prefix: "/entry"
						origElem := elem
						if l := len("/entry"); len(elem) >= l && elem[0:l] == "/entry" {
							elem = elem[l:]
						} else {
							break
						}

						if len(elem) == 0 {
							switch method {
							case "GET":
								r.name = PipelineEntryListOperation
								r.summary = ""
								r.operationID = "pipeline-entry-list"
								r.pathPattern = "/pipeline/{uuid}/entry"
								r.args = args
								r.count = 1
								return r, true
							case "POST":
								r.name = PipelineEntryCreateOperation
								r.summary = ""
								r.operationID = "pipeline-entry-create"
								r.pathPattern = "/pipeline/{uuid}/entry"
								r.args = args
								r.count = 1
								return r, true
							default:
								return
							}
						}
						switch elem[0] {
						case '/': // Prefix: "/"
							origElem := elem
							if l := len("/"); len(elem) >= l && elem[0:l] == "/" {
								elem = elem[l:]
							} else {
								break
							}

							// Param: "entry_uuid"
							// Leaf parameter
							args[1] = elem
							elem = ""

							if len(elem) == 0 {
								// Leaf node.
								switch method {
								case "DELETE":
									r.name = PipelineEntryDeleteOperation
									r.summary = ""
									r.operationID = "pipeline-entry-delete"
									r.pathPattern = "/pipeline/{uuid}/entry/{entry_uuid}"
									r.args = args
									r.count = 2
									return r, true
								case "GET":
									r.name = PipelineEntryGetOperation
									r.summary = ""
									r.operationID = "pipeline-entry-get"
									r.pathPattern = "/pipeline/{uuid}/entry/{entry_uuid}"
									r.args = args
									r.count = 2
									return r, true
								case "PUT":
									r.name = PipelineEntryUpdateOperation
									r.summary = ""
									r.operationID = "pipeline-entry-update"
									r.pathPattern = "/pipeline/{uuid}/entry/{entry_uuid}"
									r.args = args
									r.count = 2
									return r, true
								default:
									return
								}
							}

							elem = origElem
						}

						elem = origElem
					}

					elem = origElem
				}

				elem = origElem
			case 's': // Prefix: "storage"
				origElem := elem
				if l := len("storage"); len(elem) >= l && elem[0:l] == "storage" {
					elem = elem[l:]
				} else {
					break
				}

				if len(elem) == 0 {
					switch method {
					case "GET":
						r.name = StorageListOperation
						r.summary = ""
						r.operationID = "storage-list"
						r.pathPattern = "/storage"
						r.args = args
						r.count = 0
						return r, true
					default:
						return
					}
				}
				switch elem[0] {
				case '/': // Prefix: "/postgres"
					origElem := elem
					if l := len("/postgres"); len(elem) >= l && elem[0:l] == "/postgres" {
						elem = elem[l:]
					} else {
						break
					}

					if len(elem) == 0 {
						switch method {
						case "POST":
							r.name = StoragePostgresCreateOperation
							r.summary = ""
							r.operationID = "storage-postgres-create"
							r.pathPattern = "/storage/postgres"
							r.args = args
							r.count = 0
							return r, true
						default:
							return
						}
					}
					switch elem[0] {
					case '/': // Prefix: "/"
						origElem := elem
						if l := len("/"); len(elem) >= l && elem[0:l] == "/" {
							elem = elem[l:]
						} else {
							break
						}

						// Param: "uuid"
						// Leaf parameter
						args[0] = elem
						elem = ""

						if len(elem) == 0 {
							// Leaf node.
							switch method {
							case "DELETE":
								r.name = StoragePostgresDeleteOperation
								r.summary = ""
								r.operationID = "storage-postgres-delete"
								r.pathPattern = "/storage/postgres/{uuid}"
								r.args = args
								r.count = 1
								return r, true
							case "GET":
								r.name = StoragePostgresGetOperation
								r.summary = ""
								r.operationID = "storage-postgres-get"
								r.pathPattern = "/storage/postgres/{uuid}"
								r.args = args
								r.count = 1
								return r, true
							case "PUT":
								r.name = StoragePostgresUpdateOperation
								r.summary = ""
								r.operationID = "storage-postgres-update"
								r.pathPattern = "/storage/postgres/{uuid}"
								r.args = args
								r.count = 1
								return r, true
							default:
								return
							}
						}

						elem = origElem
					}

					elem = origElem
				}

				elem = origElem
			case 't': // Prefix: "tg"
				origElem := elem
				if l := len("tg"); len(elem) >= l && elem[0:l] == "tg" {
					elem = elem[l:]
				} else {
					break
				}

				if len(elem) == 0 {
					// Leaf node.
					switch method {
					case "GET":
						r.name = TgSessionListOperation
						r.summary = ""
						r.operationID = "tg-session-list"
						r.pathPattern = "/tg"
						r.args = args
						r.count = 0
						return r, true
					case "POST":
						r.name = TgSessionCreateOperation
						r.summary = ""
						r.operationID = "tg-session-create"
						r.pathPattern = "/tg"
						r.args = args
						r.count = 0
						return r, true
					case "PUT":
						r.name = TgSessionVerifyOperation
						r.summary = ""
						r.operationID = "tg-session-verify"
						r.pathPattern = "/tg"
						r.args = args
						r.count = 0
						return r, true
					default:
						return
					}
				}

				elem = origElem
			}

			elem = origElem
		}
	}
	return r, false
}
