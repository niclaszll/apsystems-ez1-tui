package apsystems

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

func (c *Client) GetDeviceInfo(ctx context.Context) (*DeviceInfo, error) {
	var info DeviceInfo
	err := c.doRequest(ctx, http.MethodGet, "/getDeviceInfo", nil, &info)
	if err != nil {
		return nil, fmt.Errorf("get device info: %w", err)
	}
	return &info, nil
}

func (c *Client) GetAlarmInfo(ctx context.Context) (*AlarmInfo, error) {
	var alarm AlarmInfo
	err := c.doRequest(ctx, http.MethodGet, "/getAlarm", nil, &alarm)
	if err != nil {
		return nil, fmt.Errorf("get alarm info: %w", err)
	}
	return &alarm, nil
}

func (c *Client) GetOutputData(ctx context.Context) (*OutputData, error) {
	var output OutputData
	err := c.doRequest(ctx, http.MethodGet, "/getOutputData", nil, &output)
	if err != nil {
		return nil, fmt.Errorf("get output data: %w", err)
	}
	return &output, nil
}

func (c *Client) GetMaxPower(ctx context.Context) (*PowerLimit, error) {
	var limit PowerLimit
	err := c.doRequest(ctx, http.MethodGet, "/getMaxPower", nil, &limit)
	if err != nil {
		return nil, fmt.Errorf("get max power: %w", err)
	}
	return &limit, nil
}

func (c *Client) SetMaxPower(ctx context.Context, watts int) error {
	if watts < 30 || watts > 800 {
		return fmt.Errorf("power must be between 30 and 800 watts, got %d", watts)
	}

	var resp PowerLimit
	endpoint := fmt.Sprintf("/setMaxPower?p=%d", watts)
	err := c.doRequest(ctx, http.MethodGet, endpoint, nil, &resp)
	if err != nil {
		return fmt.Errorf("set max power: %w", err)
	}
	return nil
}

func (c *Client) GetDevicePowerStatus(ctx context.Context) (*PowerStatus, error) {
	var status PowerStatus
	err := c.doRequest(ctx, http.MethodGet, "/getOnOff", nil, &status)
	if err != nil {
		return nil, fmt.Errorf("get power status: %w", err)
	}
	return &status, nil
}

func (c *Client) SetDevicePowerStatus(ctx context.Context, status string) error {
	var statusCode int
	switch status {
	case "ON", "NORMAL":
		statusCode = 0
	case "OFF":
		statusCode = 1
	case "SLEEP":
		statusCode = 2
	default:
		return fmt.Errorf("invalid status: %s (must be ON, OFF, or SLEEP)", status)
	}

	var resp PowerStatus
	endpoint := fmt.Sprintf("/setOnOff?status=%d", statusCode)
	err := c.doRequest(ctx, http.MethodGet, endpoint, nil, &resp)
	if err != nil {
		return fmt.Errorf("set power status: %w", err)
	}
	return nil
}

func (c *Client) GetStatistics(ctx context.Context) (*Statistics, error) {
	output, err := c.GetOutputData(ctx)
	if err != nil {
		return nil, err
	}

	return &Statistics{
		TotalPower:          output.Data.P1 + output.Data.P2,
		TotalEnergyToday:    output.Data.E1 + output.Data.E2,
		TotalEnergyLifetime: output.Data.Te1 + output.Data.Te2,
		LastUpdate:          time.Now(),
	}, nil
}
