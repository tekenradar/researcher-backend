package utils

import (
	"errors"
	"strconv"
	"strings"

	"github.com/influenzanet/study-service/pkg/types"
)

func FindSurveyItemResponse(response []types.SurveyItemResponse, itemKey string) (types.SurveyItemResponse, error) {
	for _, resp := range response {
		if strings.Contains(resp.Key, itemKey) {
			return resp, nil
		}
	}
	return types.SurveyItemResponse{}, errors.New("Could not find response item")
}

func FindResponseSlot(rootItem *types.ResponseItem, slotKey string) (*types.ResponseItem, error) {
	keyParts := strings.Split(slotKey, ".")
	if len(keyParts) > 1 {
		for _, item := range rootItem.Items {
			res, err := FindResponseSlot(&item, strings.Join(keyParts[1:], "."))
			if err == nil {
				return res, nil
			}
		}
	} else {
		if rootItem.Key == keyParts[0] {
			return rootItem, nil
		}
	}

	return nil, errors.New("could not find response slot")
}

func ExtractResponseValue(responses []types.SurveyItemResponse, itemKey string, slotKey string) (string, error) {
	surveyItem, err := FindSurveyItemResponse(responses, itemKey)
	if err != nil {
		return "", err
	}

	slotResponse, err := FindResponseSlot(surveyItem.Response, slotKey)
	if err != nil {
		return "", err
	}

	return slotResponse.Value, nil
}

func ExtractResponseValueAsNum(responses []types.SurveyItemResponse, itemKey string, slotKey string) (int64, error) {
	surveyItem, err := FindSurveyItemResponse(responses, itemKey)
	if err != nil {
		return 0, err
	}

	slotResponse, err := FindResponseSlot(surveyItem.Response, slotKey)
	if err != nil {
		return 0, err
	}

	val, err := strconv.Atoi(slotResponse.Value)
	if err != nil {
		return 0, err
	}
	return int64(val), nil
}

func MapSingleChoiceResponse(responses []types.SurveyItemResponse, itemKey string, mapping map[string]string) (string, error) {
	surveyItem, err := FindSurveyItemResponse(responses, itemKey)
	if err != nil {
		return "", err
	}

	slotResponse, err := FindResponseSlot(surveyItem.Response, "rg.scg")
	if err != nil {
		return "", err
	}

	if len(slotResponse.Items) < 1 {
		return "", errors.New("no response found")
	}

	value, ok := mapping[slotResponse.Items[0].Key]
	if !ok {
		return "", errors.New("unknown response")
	}

	return value, nil
}
