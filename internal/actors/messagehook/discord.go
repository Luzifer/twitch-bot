package messagehook

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
	"gopkg.in/irc.v4"

	"github.com/Luzifer/go_helpers/fieldcollection"
	"github.com/Luzifer/twitch-bot/v3/internal/helpers"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

type (
	discordActor struct {
		plugins.ActorKit
	}

	discordPayload struct {
		Content   string                `json:"content"`
		Username  string                `json:"username,omitempty"`
		AvatarURL string                `json:"avatar_url,omitempty"`
		Embeds    []discordPayloadEmbed `json:"embeds,omitempty"`
	}

	discordPayloadEmbed struct {
		Title       string                     `json:"title,omitempty"`
		Description string                     `json:"description,omitempty"`
		URL         string                     `json:"url,omitempty"`
		Color       int64                      `json:"color,omitempty"`
		Image       *discordPayloadEmbedImage  `json:"image,omitempty"`
		Thumbnail   *discordPayloadEmbedImage  `json:"thumbnail,omitempty"`
		Author      *discordPayloadEmbedAuthor `json:"author,omitempty"`
		Fields      []discordPayloadEmbedField `json:"fields,omitempty"`
	}

	discordPayloadEmbedAuthor struct {
		Name    string `json:"name"`
		URL     string `json:"url,omitempty"`
		IconURL string `json:"icon_url,omitempty"`
	}

	discordPayloadEmbedField struct {
		Name   string `json:"name"`
		Value  string `json:"value"`
		Inline bool   `json:"inline"`
	}

	discordPayloadEmbedImage struct {
		URL string `json:"url"`
	}
)

func (d discordActor) Execute(_ *irc.Client, m *irc.Message, r *plugins.Rule, eventData *fieldcollection.FieldCollection, attrs *fieldcollection.FieldCollection) (preventCooldown bool, err error) {
	var payload discordPayload

	if payload.Content, err = formatMessage(attrs.MustString("content", helpers.Ptr("")), m, r, eventData); err != nil {
		return false, errors.Wrap(err, "parsing content")
	}

	if payload.Username, err = formatMessage(attrs.MustString("username", helpers.Ptr("")), m, r, eventData); err != nil {
		return false, errors.Wrap(err, "parsing username")
	}

	if payload.AvatarURL, err = formatMessage(attrs.MustString("avatar_url", helpers.Ptr("")), m, r, eventData); err != nil {
		return false, errors.Wrap(err, "parsing avatar_url")
	}

	if err = d.addEmbed(&payload, m, r, eventData, attrs); err != nil {
		return false, err
	}

	return sendPayload(attrs.MustString("hook_url", helpers.Ptr("")), payload, http.StatusNoContent)
}

func (discordActor) IsAsync() bool { return false }

func (discordActor) Name() string { return "discordhook" }

func (d discordActor) Validate(tplValidator plugins.TemplateValidatorFunc, attrs *fieldcollection.FieldCollection) (err error) {
	if err = d.ValidateRequireNonEmpty(attrs, "hook_url"); err != nil {
		return err //nolint:wrapcheck
	}

	if err = d.ValidateRequireValidTemplate(tplValidator, attrs, "content"); err != nil {
		return err //nolint:wrapcheck
	}

	if err = d.ValidateRequireValidTemplateIfSet(tplValidator, attrs, "avatar_url", "username"); err != nil {
		return err //nolint:wrapcheck
	}

	if !attrs.MustBool("add_embed", helpers.Ptr(false)) {
		// We're not validating the rest if embeds are disabled but in
		// this case the content is mandatory
		return d.ValidateRequireNonEmpty(attrs, "content") //nolint:wrapcheck
	}

	//nolint:wrapcheck
	return d.ValidateRequireValidTemplateIfSet(
		tplValidator, attrs,
		"embed_title",
		"embed_description",
		"embed_url",
		"embed_image",
		"embed_thumbnail",
		"embed_author_name",
		"embed_author_url",
		"embed_author_icon_url",
		"embed_fields",
	)
}

//nolint:gocyclo // It's complex but just a bunch of converters
func (discordActor) addEmbed(payload *discordPayload, m *irc.Message, r *plugins.Rule, eventData *fieldcollection.FieldCollection, attrs *fieldcollection.FieldCollection) (err error) {
	if !attrs.MustBool("add_embed", helpers.Ptr(false)) {
		// No embed? No problem!
		return nil
	}

	var (
		embed discordPayloadEmbed
		sv    string
	)

	if embed.Title, err = formatMessage(attrs.MustString("embed_title", helpers.Ptr("")), m, r, eventData); err != nil {
		return errors.Wrap(err, "parsing embed_title")
	}

	if embed.Description, err = formatMessage(attrs.MustString("embed_description", helpers.Ptr("")), m, r, eventData); err != nil {
		return errors.Wrap(err, "parsing embed_description")
	}

	if embed.URL, err = formatMessage(attrs.MustString("embed_url", helpers.Ptr("")), m, r, eventData); err != nil {
		return errors.Wrap(err, "parsing embed_url")
	}

	if sv, err = formatMessage(attrs.MustString("embed_image", helpers.Ptr("")), m, r, eventData); err != nil {
		return errors.Wrap(err, "parsing embed_image")
	} else if sv != "" {
		embed.Image = &discordPayloadEmbedImage{URL: sv}
	}

	if sv, err = formatMessage(attrs.MustString("embed_thumbnail", helpers.Ptr("")), m, r, eventData); err != nil {
		return errors.Wrap(err, "parsing embed_thumbnail")
	} else if sv != "" {
		embed.Thumbnail = &discordPayloadEmbedImage{URL: sv}
	}

	if sv, err = formatMessage(attrs.MustString("embed_author_name", helpers.Ptr("")), m, r, eventData); err != nil {
		return errors.Wrap(err, "parsing embed_author_name")
	} else if sv != "" {
		embed.Author = &discordPayloadEmbedAuthor{Name: sv}

		if embed.Author.URL, err = formatMessage(attrs.MustString("embed_author_url", helpers.Ptr("")), m, r, eventData); err != nil {
			return errors.Wrap(err, "parsing embed_author_url")
		}

		if embed.Author.IconURL, err = formatMessage(attrs.MustString("embed_author_icon_url", helpers.Ptr("")), m, r, eventData); err != nil {
			return errors.Wrap(err, "parsing embed_author_icon_url")
		}
	}

	if sv, err = formatMessage(attrs.MustString("embed_fields", helpers.Ptr("")), m, r, eventData); err != nil {
		return errors.Wrap(err, "parsing embed_fields")
	} else if sv != "" {
		var flds []discordPayloadEmbedField
		if err = json.Unmarshal([]byte(sv), &flds); err != nil {
			return errors.Wrap(err, "unmarshalling embed_fields")
		}

		embed.Fields = flds
	}

	payload.Embeds = append(payload.Embeds, embed)
	return nil
}

//nolint:funlen // This is just a bunch of field descriptions
func (discordActor) register(args plugins.RegistrationArguments) {
	args.RegisterActor("discordhook", func() plugins.Actor { return &discordActor{} })

	args.RegisterActorDocumentation(plugins.ActionDocumentation{
		Description: "Sends a message to a Discord Web-hook",
		Name:        "Discord Message-Webhook",
		Type:        "discordhook",

		Fields: []plugins.ActionDocumentationField{
			{
				Description:     "URL to send the POST request to",
				Key:             "hook_url",
				Name:            "Hook URL",
				Optional:        false,
				SupportTemplate: false,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
			{
				Description:     "Overwrites the username set in the webhook configuration",
				Key:             "username",
				Name:            "Username",
				Optional:        true,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
			{
				Description:     "Overwrites the avatar set in the webhook configuration",
				Key:             "avatar_url",
				Name:            "Avatar URL",
				Optional:        true,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
			{
				Description:     "Message content to send to the web-hook (this must be set if embed is disabled)",
				Key:             "content",
				Name:            "Message",
				Optional:        true,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
			{
				Default:         "false",
				Description:     "Whether to include the embed in the post",
				Key:             "add_embed",
				Name:            "Add Embed",
				Optional:        true,
				SupportTemplate: false,
				Type:            plugins.ActionDocumentationFieldTypeBool,
			},
			{
				Description:     "Title of the embed",
				Key:             "embed_title",
				Name:            "Embed Title",
				Optional:        true,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
			{
				Description:     "Description of the embed",
				Key:             "embed_description",
				Name:            "Embed Description",
				Optional:        true,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
			{
				Description:     "URL the title should link to",
				Key:             "embed_url",
				Name:            "Embed URL",
				Optional:        true,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
			{
				Description:     "URL of the big image displayed in the embed",
				Key:             "embed_image",
				Name:            "Embed Image URL",
				Optional:        true,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
			{
				Description:     "URL of the small image displayed in the embed",
				Key:             "embed_thumbnail",
				Name:            "Embed Thumbnail URL",
				Optional:        true,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
			{
				Description:     "Name of the post author (if empty all other author-fields are ignored)",
				Key:             "embed_author_name",
				Name:            "Embed Author Name",
				Optional:        true,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
			{
				Description:     "URL the author name should link to",
				Key:             "embed_author_url",
				Name:            "Embed Author URL",
				Optional:        true,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
			{
				Description:     "URL of the author avatar",
				Key:             "embed_author_icon_url",
				Name:            "Embed Author Avatar URL",
				Optional:        true,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
			{
				Description:     "Fields to display in the embed (must yield valid JSON: `[{\"name\": \"\", \"value\": \"\", \"inline\": false}]`)",
				Key:             "embed_fields",
				Name:            "Embed Fields",
				Optional:        true,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
		},
	})
}
