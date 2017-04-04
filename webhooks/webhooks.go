package webhooks

import (
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/blacklightcms/recurly"
)

// Webhook notification constants.
const (
	NewAccount         = "new_account_notification"
	CanceledAccount    = "cancelled_account_notification"
	ReactivatedAccount = "reactivated_account_notification"

	NewSubscription      = "new_subscription_notification"
	UpdatedSubscription  = "updated_subscription_notification"
	CanceledSubscription = "canceled_subscription_notification"
	ExpiredSubscription  = "expired_subscription_notification"
	RenewedSubscription  = "renewed_subscription_notification"

	NewInvoice        = "new_invoice_notification"
	ProcessingInvoice = "processing_invoice_notification"
	ClosedInvoice     = "closed_invoice_notification"
	PastDueInvoice    = "past_due_invoice_notification"

	SuccessfulPayment = "successful_payment_notification"
	FailedPayment     = "failed_payment_notification"
)

type notificationName struct {
	XMLName xml.Name
}

type (
	// NewAccountNotification is sent when a new account is created
	NewAccountNotification struct {
		Account recurly.Account `xml:"account"`
	}

	// CanceledAccountNotification is sent when an account is closed
	CanceledAccountNotification struct {
		Account recurly.Account `xml:"account"`
	}

	// ReactivatedAccountNotification is sent when an account subscription is reactivated after having been canceled
	ReactivatedAccountNotification struct {
		Account recurly.Account `xml:"account"`
	}

	// NewSubscriptionNotification is sent when a new subscription is created
	NewSubscriptionNotification struct {
		Account      recurly.Account      `xml:"account"`
		Subscription recurly.Subscription `xml:"subscription"`
	}

	// UpdatedSubscriptionNotification is sent when a subscription is upgraded or downgraded
	UpdatedSubscriptionNotification struct {
		Account      recurly.Account      `xml:"account"`
		Subscription recurly.Subscription `xml:"subscription"`
	}

	// CanceledSubscriptionNotification is sent when a subscription is canceled
	CanceledSubscriptionNotification struct {
		Account      recurly.Account      `xml:"account"`
		Subscription recurly.Subscription `xml:"subscription"`
	}

	// ExpiredSubscriptionNotification is sent when a subscription is no longer valid
	ExpiredSubscriptionNotification struct {
		Account      recurly.Account      `xml:"account"`
		Subscription recurly.Subscription `xml:"subscription"`
	}

	// RenewedSubscriptionNotification is sent whenever a subscription renews
	RenewedSubscriptionNotification struct {
		Account      recurly.Account      `xml:"account"`
		Subscription recurly.Subscription `xml:"subscription"`
	}

	// SuccessfulPaymentNotification is sent when a payment is successful.
	SuccessfulPaymentNotification struct {
		Account     recurly.Account     `xml:"account"`
		Transaction recurly.Transaction `xml:"transaction"`
	}

	// FailedPaymentNotification is sent when a payment fails.
	FailedPaymentNotification struct {
		Account     recurly.Account     `xml:"account"`
		Transaction recurly.Transaction `xml:"transaction"`
	}

	// NewInvoiceNotification is sent when a new invoice is generated.
	NewInvoiceNotification struct {
		Account recurly.Account `xml:"account"`
		Invoice recurly.Invoice `xml:"invoice"`
	}

	// ProcessingInvoiceNotification is sent when a new invoice is generated.
	ProcessingInvoiceNotification struct {
		Account recurly.Account `xml:"account"`
		Invoice recurly.Invoice `xml:"invoice"`
	}

	// ClosedInvoiceNotification is sent when a new invoice is generated.
	ClosedInvoiceNotification struct {
		Account recurly.Account `xml:"account"`
		Invoice recurly.Invoice `xml:"invoice"`
	}

	// PastDueInvoiceNotification is sent when an invoice is past due.
	PastDueInvoiceNotification struct {
		Account recurly.Account `xml:"account"`
		Invoice recurly.Invoice `xml:"invoice"`
	}
)

// transactionHolder allows the uuid and invoice number fields to be set.
// The UUID field is labeled id in notifications and the invoice number
// is not included on the existing transaction struct.
type transactionHolder interface {
	setTransactionFields(id string, in int)
}

// setTransactionFields sets fields on the transaction struct.
func (n *SuccessfulPaymentNotification) setTransactionFields(id string, invoiceNumber int) {
	n.Transaction.UUID = id
	n.Transaction.InvoiceNumber = invoiceNumber
}

func (n *FailedPaymentNotification) setTransactionFields(id string, invoiceNumber int) {
	n.Transaction.UUID = id
	n.Transaction.InvoiceNumber = invoiceNumber
}

// transaction allows the transaction id and invoice number to be unmarshalled
// so they can be set on the notification struct.
type transaction struct {
	ID            string `xml:"transaction>id"`
	InvoiceNumber int    `xml:"transaction>invoice_number,omitempty"`
}

// ErrUnknownNotification is used when the incoming webhook does not match a
// predefined notification type. It implements the error interface.
type ErrUnknownNotification struct {
	name string
}

// Error implements the error interface.
func (e ErrUnknownNotification) Error() string {
	return fmt.Sprintf("unknown notification: %s", e.name)
}

// Name returns the name of the unknown notification.
func (e ErrUnknownNotification) Name() string {
	return e.name
}

// Parse parses an incoming webhook and returns the notification.
func Parse(r io.Reader) (interface{}, error) {
	if closer, ok := r.(io.Closer); ok {
		defer closer.Close()
	}

	notification, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var n notificationName
	if err := xml.Unmarshal(notification, &n); err != nil {
		return nil, err
	}

	var dst interface{}
	switch n.XMLName.Local {
	case NewAccount:
		dst = &NewAccountNotification{}
	case CanceledAccount:
		dst = &CanceledAccountNotification{}
	case ReactivatedAccount:
		dst = &ReactivatedAccountNotification{}
	case NewSubscription:
		dst = &NewSubscriptionNotification{}
	case UpdatedSubscription:
		dst = &UpdatedSubscriptionNotification{}
	case CanceledSubscription:
		dst = CanceledSubscriptionNotification{}
	case ExpiredSubscription:
		dst = &ExpiredSubscriptionNotification{}
	case RenewedSubscription:
		dst = &RenewedSubscriptionNotification{}
	case NewInvoice:
		dst = &NewInvoiceNotification{}
	case ProcessingInvoice:
		dst = &ProcessingInvoiceNotification{}
	case ClosedInvoice:
		dst = &ClosedInvoiceNotification{}
	case PastDueInvoice:
		dst = &PastDueInvoiceNotification{}
	case SuccessfulPayment:
		dst = &SuccessfulPaymentNotification{}
	case FailedPayment:
		dst = &FailedPaymentNotification{}
	default:
		return nil, ErrUnknownNotification{name: n.XMLName.Local}
	}

	if err := xml.Unmarshal(notification, dst); err != nil {
		return nil, err
	}

	if th, ok := dst.(transactionHolder); ok {
		var t transaction
		if err := xml.Unmarshal(notification, &t); err != nil {
			return nil, err
		}
		th.setTransactionFields(t.ID, t.InvoiceNumber)
	}

	return dst, nil
}
