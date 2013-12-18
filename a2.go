package main

import (
    "fmt"
    "sync"
    "time"
    "math"
    "math/rand"
)

type TermScore struct {
    Term string
    Score float64
}

type VitalStats struct {
    N int
    Mean float64
    SD float64
}

func main() {
    // first let's seed the random number generator properly
    rand.Seed( time.Now().UTC().UnixNano())

    var scores_to_generate int = 10
    var words []string = []string{
        "magniloquent",
        "hetorodoxia",
        "fulgurant",
        "limpid",
        "sod",
    }

    var _audit map [string] []float64 = make(map [string] []float64)
    var leader_board map [string] VitalStats = make(map [string] VitalStats)

    var processing map [string] int = make(map [string] int)
    var next_to_process map [string] int = make(map [string] int)
    var i int = 0

    var c chan TermScore = make(chan TermScore)
    go TermScores(c, scores_to_generate, words)

    var wg sync.WaitGroup
    for term_score := range c {
        // break on empty string - SHOULD use ok or err instead but this works for our purposes
        if term_score.Term == "" { break }

        var term string = term_score.Term
        var score float64 = term_score.Score

        // first append it to the audit group (synchronously)
        if _, exists := _audit[term]; !exists {
            _audit[term] = []float64{score}
        } else {
            _audit[term] = append(_audit[term], score)
        }

        // then handle the online formulae (asynchronously)
        wg.Add(1)
        go func(id int) {
            defer wg.Done()

            // Preprotocol
            for {
                current, current_exists := processing[term]

                if current_exists {
                    if current == id {
                        break  // "I am now processing this term!"
                    }  // "Someone else is currently processing this term - I will try again later"
                } else {  // "No-one is currently processing this term!"
                    next, next_exists := next_to_process[term]
                    if next_exists {  // "Someone is next to process this term, so..."
                        processing[term] = next  // "...whoever it is is now processing this term."
                        if next != id {  // "If it's not me..."
                            next_to_process[term] = id  // "... then I am next to process this term!"
                        } else {  // "I am now processing this term!"
                            delete(next_to_process, term)  // "Someone else is next to process this term."
                        }
                    } else {  // "No-one is next to process this term!"
                        next_to_process[term] = id  // "I am next to process this term!"
                    }
                }

                // random sleep to keep things from overheating
                duration := time.Duration(rand.Intn(100)) * time.Millisecond  // some fraction of a centisecond
                time.Sleep(duration)  // during which time other reads/writes happen
            }

            var new_n int = 1
            var new_mean float64 = score
            var new_sd float64 = 0.0
            
            if term_stats, exists := leader_board[term]; exists {
                // Math time!

                // STEP 1: the count
                // aka k
                old_n := term_stats.N
                new_n = old_n + 1

                // STEP 2: the mean
                // M(k) = M(k-1) + (x(k) - M(k-1)) / k
                old_mean := term_stats.Mean
                new_mean = old_mean + (score - old_mean) / float64(new_n)  // Knuth-Welford

                // STEP 3: the standard deviation
                // S(k) = S(k-1) + (x(k) - M(k-1)) * (x(k) - M(k))
                old_sd := term_stats.SD
                sum_of_squared_differences := old_sd * old_sd * float64(old_n)
                sum_of_squared_differences += (score - old_mean) * (score - new_mean)  // Knuth-Welford
                new_sd = math.Sqrt(sum_of_squared_differences / float64(new_n))

                // TODO: HERE IS WHERE YOU WOULD DO CORRELATION
                // old_covariance := old_correlation * old_sd_y * old_sd_x 
                // new_covariance := (old_covariance * n + (score_x - new_mean_x) * (score_y - old_mean_y)) / n  // Pébay
                // new_correlation := new_covariance / (new_sd_y * new_sd_x)
            }

            // here's where things go haywire with the concurrency
            duration := time.Duration(rand.Intn(100)) * time.Millisecond  // some fraction of a centisecond
            time.Sleep(duration)  // during which time other reads/writes happen

            // update/create vital stats for term
            leader_board[term] = VitalStats{
                new_n,
                new_mean,
                new_sd,
            }

            // "I am done processing!""
            delete(processing, term)
        } (i)
        i++
    }

    wg.Wait()

    fmt.Printf("Processed %d term scores\n", scores_to_generate)

    fmt.Printf("\nSynchronous:\n")
    for term, scores := range _audit {
        n := len(scores)
        total := 0.0
        for _, score := range scores {
            total += score
        }
        mean := total / float64(n)
        
        sum_of_squared_differences := 0.0
        for _, score := range scores {
            sum_of_squared_differences += (score - mean) * (score - mean)
        }
        sd := math.Sqrt(sum_of_squared_differences / float64(n))

        fmt.Printf("%s: mean=%f; sd=%f\n", term, mean, sd)
    }

    fmt.Printf("\nAsynchronous:\n")
    for term, stats := range leader_board {
        fmt.Printf("%s: mean=%f; sd=%f\n", term, stats.Mean, stats.SD)
    }
}

func TermScores(c chan TermScore, n int, words []string) {
    // generates a bunch of pseudo-random data
    for i := 0; i < n; i++ {
        word := words[rand.Intn(len(words))]
        score := rand.Float64()
        c <- TermScore{word, score}
    }

    c <- TermScore{"", 0.0}
}

