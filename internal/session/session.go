package session

import (
	"context"
	"fmt"
	"strings"

	"cybros/consts"
	"cybros/internal/logger"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
	"github.com/sirupsen/logrus"
)

type Session struct {
	client *telegram.Client
	auth   auth.UserAuthenticator
}

func New(client *telegram.Client, userAuth auth.UserAuthenticator) Session {
	return Session{
		client: client,
		auth:   userAuth,
	}
}

func (s Session) Run(ctx context.Context) error {
	loginErr := s.client.Auth().IfNecessary(ctx, auth.NewFlow(s.auth, auth.SendCodeOptions{}))
	if loginErr != nil {
		return fmt.Errorf(consts.ErrorInitializeLogin, loginErr)
	}

	self, err := s.client.Self(ctx)
	if err != nil {
		return fmt.Errorf(consts.ErrorGetSelf, err)
	}
	logAccount(self)

	<-ctx.Done()
	return nil
}

func logAccount(self *tg.User) {
	fields := logrus.Fields{
		"id":       self.ID,
		"bot":      self.Bot,
		"premium":  self.Premium,
		"verified": self.Verified,
	}

	firstName, _ := self.GetFirstName()
	lastName, _ := self.GetLastName()
	name := strings.TrimSpace(firstName + " " + lastName)
	if name != "" {
		fields["name"] = name
	}
	if firstName != "" {
		fields["first_name"] = firstName
	}
	if lastName != "" {
		fields["last_name"] = lastName
	}
	username, ok := self.GetUsername()
	if ok && username != "" {
		fields["username"] = username
	}
	langCode, ok := self.GetLangCode()
	if ok && langCode != "" {
		fields["lang_code"] = langCode
	}

	logger.Log.WithFields(fields).Info("telegram account")
}
