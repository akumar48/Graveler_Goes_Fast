package main

import (
    "fmt"
    "math/rand"
    "sync"
    "time"
    _ "net/http/pprof" // Import for side effects
    "net/http"
    "log"
)

const rollsPerSession = 231
const maxRolls = 1000000000
const maxTarget = 177
const numWorkers = 8 // Number of parallel workers (tune this to match your CPU cores)

func rollSessionOnlyOnes(r *rand.Rand) int {
    countOnes := 0
    for i := 0; i < rollsPerSession; i++ {
        roll := r.Intn(4) // Use the local random generator instance
        if roll == 0 {    // Only count the "1" rolls
            countOnes++
        }
    }
    return countOnes
}

func worker(id int, wg *sync.WaitGroup, maxOnes *int, rolls *int, mu *sync.Mutex, resultChan chan int) {
    defer wg.Done()

    localRand := rand.New(rand.NewSource(time.Now().UnixNano() + int64(id))) // Create a local rand instance
    localMaxOnes := 0
    localRolls := 0

    for localRolls < maxRolls / numWorkers { // Divide work among workers
        count := rollSessionOnlyOnes(localRand)
        if count > localMaxOnes {
            localMaxOnes = count
        }
        localRolls++
        if localMaxOnes >= maxTarget {
            break
        }
    }

    // Update shared variables in a thread-safe manner
    mu.Lock()
    if localMaxOnes > *maxOnes {
        *maxOnes = localMaxOnes
    }
    *rolls += localRolls
    mu.Unlock()

    resultChan <- localMaxOnes
}

func main() {
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()

    start := time.Now()

    var maxOnes int
    var totalRolls int
    var mu sync.Mutex
    var wg sync.WaitGroup
    resultChan := make(chan int, numWorkers)

    // Start workers
    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go worker(i, &wg, &maxOnes, &totalRolls, &mu, resultChan)
    }

    wg.Wait()
    close(resultChan)

    duration := time.Since(start)
    fmt.Println("Highest Ones Roll:", maxOnes)
    fmt.Println("Number of Roll Sessions:", totalRolls)
    fmt.Println("Duration:", duration)
}
