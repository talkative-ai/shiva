package routes

import (
	"fmt"
	"net/http"
	"os"
	"time"

	stripe "github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/client"
	"github.com/talkative-ai/core/prehandle"
	"github.com/talkative-ai/core/router"
)

// PostSubscribe router.Route
/* Path: "/workbench/v1/auth/google"
 * Method: "GET"
 * Validates a Google OAuth 2 token.
 * Responds with a hashmap containing an intercom.io token
 *		and models.User model
 */
var PostSubscribe = &router.Route{
	Path:       "/workbench/v1/subscribe",
	Method:     "POST",
	Handler:    http.HandlerFunc(postSubscribeHandler),
	Prehandler: []prehandle.Prehandler{},
}

type postSubscribePayload struct {
	Token      string
	GivenName  string
	FamilyName string
}

func postSubscribeHandler(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		fmt.Println("Error", err.Error())
	}

	var plan string

	if os.Getenv("DEVELOPMENT_ENVIRONMENT") == "TESTING" {
		switch r.FormValue("plan") {
		case "full":
			plan = "plan_CkNgreTjWhFR6R"
			break
		case "unbranded":
			plan = "plan_CkNhpypR5GOQal"
			break
		default:
			plan = "plan_CkNLi5FrarJh7Z"
			break
		}
	} else {
		switch r.FormValue("plan") {
		case "full":
			plan = "plan_CkNiYEMNe4nrln"
			break
		case "unbranded":
			plan = "plan_CkNjpg3UhOHgsj"
			break
		default:
			plan = "plan_CkNhbaNN1bHUqB"
			break
		}
	}

	stripeClient := &client.API{}
	stripeClient.Init(os.Getenv("STRIPE_KEY"), nil)
	customerParams := stripe.CustomerParams{
		Email: r.FormValue("stripeEmail"),
	}
	customerParams.SetSource(r.FormValue("stripeToken"))
	customer, err := stripeClient.Customers.New(&customerParams)
	if err != nil {
		fmt.Println("Error", err.Error())
	}
	subParams := stripe.SubParams{
		Customer: customer.ID,
		Items: []*stripe.SubItemsParams{
			&stripe.SubItemsParams{
				Plan:     plan,
				Quantity: 1,
			},
		},
		TrialEnd: time.Now().AddDate(1, 0, 0).Unix(),
	}
	stripeClient.Subs.New(&subParams)

}
