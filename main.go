package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"io/ioutil"
	"strconv"
	"strings"
)

// --- consts
const eliteSavesDir = "C:\\Users\\v3133\\Saved Games\\Frontier Developments\\Elite Dangerous"
type UnstructuredJson map[string]interface{}
type HandlerFunction func (json UnstructuredJson)


// --- data structs
type ActiveMissions struct {
	MissionID float64
	Reward float64
	Commodity string
	Count float64
}

// --- data storages
var activeWingMissions = make(map[float64]*ActiveMissions) // active mission stricts indexed by the mission ids

// --- event handlers

// When Written: when starting a mission
func hMissionAccepted(json UnstructuredJson) {
	missionId := json["MissionID"].(float64)

	// check Trade Wing missions
	if _, ok := activeWingMissions[missionId]; !ok {
		if json["Commodity"] != nil && json["Reward"] != nil {

			commodityName := json["Commodity"].(string)
			commodityName = strings.ReplaceAll(commodityName, "$", "")
			commodityName = strings.ReplaceAll(commodityName, "_Name;", "")

			activeWingMissions[missionId] = &ActiveMissions{
				missionId,
				json["Reward"].(float64),
				commodityName,
				json["Count"].(float64),
			}
		}
		//fmt.Printf("MissionAccepted, %v\n", missionId)
	}

	//todo other mission handlers ...
}

// When Written: when a mission is completed
func hMissionCompleted(json UnstructuredJson) {
	missionId := json["MissionID"].(float64)
	if _, ok := activeWingMissions[missionId]; ok {
		delete(activeWingMissions, missionId)
	}
}

// When Written: when a mission has been abandoned
func hMissionAbandoned(json UnstructuredJson) {
	missionId := json["MissionID"].(float64)
	if _, ok := activeWingMissions[missionId]; ok {
		delete(activeWingMissions, missionId)
	}
}

// When Written: when a mission has failed
func hMissionFailed(json UnstructuredJson) {
	missionId := json["MissionID"].(float64)
	if _, ok := activeWingMissions[missionId]; ok {
		delete(activeWingMissions, missionId)
	}
}

// When written: when collecting or delivering cargo for a wing mission, or if a wing member updates progress
func hCargoDepot(json UnstructuredJson) {
	missionId := json["MissionID"].(float64)
	if _, ok := activeWingMissions[missionId]; ok {

		//fmt.Println(json)

		if val, ok := json["UpdateType"]; ok && val == "WingUpdate" {
			activeWingMissions[missionId].Count = json["TotalItemsToDeliver"].(float64) - json["ItemsDelivered"].(float64)
		}

		if val, ok := json["UpdateType"]; ok && val == "Deliver" {
			//fmt.Println(json)
			activeWingMissions[missionId].Count -= json["Count"].(float64)
		}

		//if json["UpdateType"].(string) == "WingUpdate" {
			//activeWingMissions[missionId].Count += json["ItemsCollected"].(float64)
			//activeWingMissions[missionId].Count -= json["ItemsDelivered"].(float64)
		//}
	}
}

func main() {
	//fmt.Println("Starting...")

	//handlers https://elite-journal.readthedocs.io
	handlers := map[string] HandlerFunction {
		"MissionAccepted": hMissionAccepted,
		"MissionCompleted": hMissionCompleted,
		"MissionAbandoned": hMissionAbandoned,
		"MissionFailed": hMissionFailed,
		"CargoDepot": hCargoDepot,
	}

	//parse each file and call handler for row if it exists
	items, _ := ioutil.ReadDir(eliteSavesDir)
	for _, item := range items {
		if strings.HasPrefix(item.Name(), "Journal") && strings.HasSuffix(item.Name(), ".log") {
			//fmt.Println(item.Name())

			inFile, _ := os.Open(eliteSavesDir + "\\" + item.Name())
			defer inFile.Close()
			scanner := bufio.NewScanner(inFile)
			scanner.Split(bufio.ScanLines)

			for scanner.Scan() {

				// optimize to prevent parse each event json
				contains := false
				for k, _ := range handlers {
					if strings.Contains(scanner.Text(), k) {
						contains = true
						break
					}
				}

				if !contains {
					continue
				}
				// end of optimize

				var result map[string]interface{}
				json.Unmarshal([]byte(scanner.Text()), &result)
				eventType := result["event"].(string)
				if _, ok := handlers[eventType]; ok {
					handlers[eventType](result)
					//fmt.Println(eventType)
				}
			}
		}
	}

	//print all active missions
	fmt.Println("Active Wing missions")
	for _, v := range activeWingMissions {
		fmt.Printf("%s, %v, %v\n", v.Commodity, v.Count, strconv.FormatFloat(v.Reward, 'f', -1, 64))
	}
	fmt.Println("")

	//calc active wing missions demand
	fmt.Println("Total Wing Demand")
	totalActiveWingMissionsDemand := make(map[string]float64)
	for _, v := range activeWingMissions {
		if _, ok := totalActiveWingMissionsDemand[v.Commodity]; ok {
			// append
			totalActiveWingMissionsDemand[v.Commodity] = totalActiveWingMissionsDemand[v.Commodity] + v.Count
		} else {
			// just insert
			totalActiveWingMissionsDemand[v.Commodity] = v.Count
		}
	}

	for k, v := range totalActiveWingMissionsDemand {
		fmt.Printf("%s = %v\n", k, v)
	}

}