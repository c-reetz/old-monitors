package main

import (
	"log"
	"monitors/mongo"
	ralph_lauren "monitors/sites/ralph-lauren"
	"monitors/sites/unisport"
	"sync"
)

func main() {
	log.SetPrefix("Logger: ")

	_, mongoErr := mongo.GetMongoClient() //for at init selve mongodb, og fange errors inden alt andet starter
	if mongoErr != nil {
		log.Fatal(mongoErr)
	}

	var wg sync.WaitGroup

	wg.Add(5)

	go func() {
		defer wg.Done()
		ralph_lauren.MensAccessory("Scandi")
	}()
	go func() {
		defer wg.Done()
		ralph_lauren.MensClothing("Scandi")
	}()
	go func() {
		defer wg.Done()
		ralph_lauren.WomensAccessory("Scandi")
	}()
	go func() {
		defer wg.Done()
		ralph_lauren.WomensClothing("Scandi")
	}()
	go func() {
		defer wg.Done()
		ralph_lauren.MensShoes("Scandi")
	}()
	go func() {
		defer wg.Done()
		ralph_lauren.WomensShoes("Scandi")
	}()
	go func() {
		defer wg.Done()
		unisport.Filtered("Scandi")
	}()
	wg.Wait()
}
