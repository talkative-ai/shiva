package routes

// PostUserRegister router.Route
// Path: "/user/register",
// Method: "POST",
// Accepts models.UserRegister
// Responds with status of success or failure
// var PostUserRegister = &router.Route{
// 	Path:       "/user/verify-email",
// 	Method:     "POST",
// 	Handler:    http.HandlerFunc(postUserRegisterHandler),
// 	Prehandler: []prehandle.Prehandler{prehandle.RequireBody(1024)},
// }

// func postUserRegisterHandler(w http.ResponseWriter, r *http.Request) {

// 	cache, err := providers.ConnectRedis(r)
// 	if err != nil {
// 		phrerrors.ServerError(w, r, err)
// 		return
// 	}

// 	defer func() {
// 		if cache != nil {
// 			cache.Close()
// 		}
// 	}()

// 	userParams := &models.StdParams{Cache: cache, W: w, R: r}
// 	var user models.User
// 	err = json.Unmarshal([]byte(r.Header.Get("X-Body")), &user)
// 	if err != nil {
// 		return
// 	}

// 	accStatus, err := user.HasAccountFlag(userParams)
// 	if err != nil {
// 		return
// 	}

// 	if accStatus&models.USER_ACCOUNT_DNE == 0 {
// 		// Account exists
// 		return
// 	}

// 	accStatus ^= models.USER_ACCOUNT_DNE
// 	accStatus |= models.USER_ACCOUNT_CREATING

// 	isNewAccount, err := user.SetAccountFlag(userParams, accStatus)
// 	if err != nil {
// 		return
// 	}

// 	if !isNewAccount {
// 		// Duplicate accounts are being created simultaneously. Abort
// 		// TODO Handle more elegantly
// 		return
// 	}

// 	user.Save(userParams)
// 	user.SendVerificationEmail(userParams)

// }
