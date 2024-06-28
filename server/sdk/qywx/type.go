package qywx

import "github.com/silenceper/wechat/v2/work/message"

// SendTextRequest 发送文本消息的请求
type TemplateCardRequest struct {
	*message.SendRequestCommon
	TemplateCard *TemplateCardButton `json:"template_card"`
}

// TemplateCardButton
// https://developer.work.weixin.qq.com/document/path/90236#%E6%A8%A1%E6%9D%BF%E5%8D%A1%E7%89%87%E6%B6%88%E6%81%AF
type TemplateCardButton struct {
	CardType              string `json:"card_type"`
	Source                `json:"source,omitempty"`
	ActionMenu            `json:"action_menu,omitempty"`
	MainTitle             `json:"main_title,omitempty"`
	QuoteArea             `json:"quote_area,omitempty"`
	SubTitleText          string              `json:"sub_title_text,omitempty"`
	HorizontalContentList []HorizontalContent `json:"horizontal_content_list,omitempty"`
	VerticalContentList   []VerticalContent   `json:"vertical_content_list,omitempty"`
	CardAction            `json:"card_action,omitempty"`
	JumpList              []Jump `json:"jump_list,omitempty"`
	EmphasisContent       `json:"emphasis_content,omitempty"`
	ImageTextArea         `json:"image_text_area,omitempty"`
	CardImage             `json:"card_image,omitempty"`
	Checkbox              `json:"checkbox,omitempty"`
	SelectList            []Select `json:"select_list,omitempty"`
	TaskId                string   `json:"task_id"`
	ButtonSelection       `json:"button_selection,omitempty"`
	ButtonList            []Button `json:"button_list,omitempty"`
}

type Source struct {
	IconUrl   string `json:"icon_url,omitempty"`
	Desc      string `json:"desc,omitempty"`
	DescColor int    `json:"desc_color,omitempty"`
}

type ActionMenu struct {
	Desc       string   `json:"desc,omitempty"`
	ActionList []Action `json:"action_list"`
}
type MainTitle struct {
	Title string `json:"title"`
	Desc  string `json:"desc,omitempty"`
}
type QuoteArea struct {
	Type      int    `json:"type,omitempty"`
	Url       string `json:"url,omitempty"`
	Title     string `json:"title,omitempty"`
	QuoteText string `json:"quote_text,omitempty"`
}
type CardAction struct {
	Type     int    `json:"type,omitempty"`
	Url      string `json:"url,omitempty"`
	Appid    string `json:"appid,omitempty"`
	Pagepath string `json:"pagepath,omitempty"`
}
type EmphasisContent struct {
	Title string `json:"title,omitempty"`
	Desc  string `json:"desc,omitempty"`
}
type ImageTextArea struct {
	Type     int    `json:"type"`
	Url      string `json:"url"`
	Title    string `json:"title"`
	Desc     string `json:"desc"`
	ImageUrl string `json:"image_url"`
}
type CardImage struct {
	Url         string  `json:"url"`
	AspectRatio float64 `json:"aspect_ratio"`
}
type Checkbox struct {
	QuestionKey string           `json:"question_key"`
	OptionList  []OptionCheckBox `json:"option_list"`
	Mode        int              `json:"mode"`
}
type ButtonSelection struct {
	QuestionKey string   `json:"question_key"`
	Title       string   `json:"title,omitempty"`
	OptionList  []Option `json:"option_list,omitempty"`
	SelectedId  string   `json:"selected_id,omitempty"`
}
type Action struct {
	Text string `json:"text"`
	Key  string `json:"key"`
}
type HorizontalContent struct {
	Keyname string `json:"keyname"`
	Value   string `json:"value,omitempty"`
	Type    int    `json:"type,omitempty"`
	Url     string `json:"url,omitempty"`
	MediaId string `json:"media_id,omitempty"`
	Userid  string `json:"userid,omitempty"`
}
type VerticalContent struct {
	Title string `json:"title"`
	Desc  string `json:"desc"`
}
type Jump struct {
	Type     int    `json:"type,omitempty"`
	Title    string `json:"title"`
	Url      string `json:"url,omitempty"`
	Appid    string `json:"appid,omitempty"`
	Pagepath string `json:"pagepath,omitempty"`
}
type OptionCheckBox struct {
	Id        string `json:"id"`
	Text      string `json:"text"`
	IsChecked bool   `json:"is_checked"`
}
type Option struct {
	Id   string `json:"id"`
	Text string `json:"text"`
}
type Select struct {
	QuestionKey string   `json:"question_key"`
	Title       string   `json:"title"`
	SelectedId  string   `json:"selected_id"`
	OptionList  []Option `json:"option_list"`
}
type Button struct {
	Text  string `json:"text"`
	Style int    `json:"style,omitempty"`
	Key   string `json:"key,omitempty"`
}
