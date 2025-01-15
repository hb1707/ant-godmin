package coze

import (
	"encoding/json"
	"errors"
	"github.com/hb1707/ant-godmin/pkg/log"
	"github.com/hb1707/exfun/fun/curl"
	"strconv"
)

type OpenSpace struct {
	Id            string `json:"id"`
	Name          string `json:"name"`
	IconUrl       string `json:"icon_url"`
	RoleType      string `json:"role_type"`
	WorkspaceType string `json:"workspace_type"`
}

type OpenSpaceData struct {
	Workspaces []OpenSpace
	TotalCount int `json:"total_count"`
}

type SpaceList struct {
	Code int           `json:"code"`
	Msg  string        `json:"msg"`
	Data OpenSpaceData `json:"data"`
}

func (c *Client) GetSpaceList() (*SpaceList, error) {
	client := curl.Config{
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer " + c.Token,
		},
	}
	if c.Page == 0 {
		c.Page = 1
	}
	params := map[string]string{
		"page_size": strconv.Itoa(20),
		"page_num":  strconv.Itoa(c.Page),
	}
	resp, _, err := client.GET(spaceListUri, params)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	var result SpaceList
	err = json.Unmarshal(resp, &result)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	if result.Code != 0 {
		log.Error(result.Msg)
		return nil, errors.New(result.Msg)
	}
	return &result, nil
}

type SpaceBot struct {
	BotId       string `json:"bot_id"`
	BotName     string `json:"bot_name"`
	Description string `json:"description"`
	IconUrl     string `json:"icon_url"`
	PublishTime string `json:"publish_time"`
}

type SpaceBots struct {
	SpaceBots []SpaceBot `json:"space_bots"`
	Total     int        `json:"total"`
}

type AgentList struct {
	Code int       `json:"code"`
	Msg  string    `json:"msg"`
	Data SpaceBots `json:"data"`
}

func (c *Client) GetAgentList(spaceID string) (*AgentList, error) {
	client := curl.Config{
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer " + c.Token,
		},
	}
	if c.Page == 0 {
		c.Page = 1
	}
	params := map[string]string{
		"space_id":   spaceID,
		"page_size":  strconv.Itoa(20),
		"page_index": strconv.Itoa(c.Page),
	}
	resp, _, err := client.GET(agentListUri, params)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	var result AgentList
	err = json.Unmarshal(resp, &result)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	if result.Code != 0 {
		log.Error(result.Msg)
		return nil, errors.New(result.Msg)
	}
	return &result, nil
}

type PublishDraftBotData struct {
	BotId   string `json:"bot_id"`
	Version string `json:"version"`
}

type RespBotPublish struct {
	Code         int                 `json:"code"`
	Msg          string              `json:"msg"`
	Data         PublishDraftBotData `json:"data"`
	Error        string              `json:"error"`
	ErrorMessage string              `json:"error_message"`
}

func (c *Client) BotPublish(botID string, connectorIds []string) (string, error) {
	//log.Debug(token)
	client := curl.Config{
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer " + c.Token,
		},
	}
	params := map[string]any{
		"bot_id":        botID,
		"connector_ids": connectorIds,
	}
	b, _ := json.Marshal(params)
	resp, _, err := client.POST(botPublish, b)
	if err != nil {
		log.Error(err)
		return "", err
	}
	var result RespBotPublish
	err = json.Unmarshal(resp, &result)
	if err != nil {
		log.Error(err)
		return "", err
	}
	if result.Error != "" {
		log.Error(result.ErrorMessage)
		return "", errors.New(result.ErrorMessage)
	}
	return result.Data.Version, nil
}
