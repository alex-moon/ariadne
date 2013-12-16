package main

import (
    "fmt"
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
    var scores_to_generate int = 10000
    var _audit map [string] []float64 = make(map [string] []float64)
    var leader_board map [string] VitalStats = make(map [string] VitalStats)

    c := make(chan TermScore)
    go TermScores(c, scores_to_generate)

    for term_score := range c {
        go func() {
            var term string = term_score.Term
            var score float64 = term_score.Score

            fmt.Printf("we got a string %s and a score %s\n", term, score)

            var new_n int = 1
            var new_mean float64 = score
            var new_sd float64 = 0.0
            
            // Math time!
            if term_stats, exists := leader_board[term]; exists {
                // STEP 1: the count
                // aka k
                old_n :=  term_stats.N
                new_n := old_n + 1

                // STEP 2: the mean
                // M(k) = M(k-1) + (x(k) - M(k-1)) / k
                old_mean := term_stats.Mean
                new_mean := old_mean + (score - old_mean) / float64(new_n)  // Knuth-Welford

                // STEP 3: the standard deviation
                // S(k) = S(k-1) + (x(k) - M(k-1)) * (x(k) - M(k))
                old_sd := term_stats.SD
                new_sd := old_sd + (score - old_mean) * (score - new_mean)  // Knuth-Welford

                // TODO: HERE IS WHERE YOU WOULD DO CORRELATION
                // old_covariance := old_correlation * old_sd_y * old_sd_x 
                // new_covariance := (old_covariance * n + (score_x - new_mean_x) * (score_y - old_mean_y)) / n  // Pébay
                // new_correlation := new_covariance / (new_sd_y * new_sd_x)
            }

            // update/create vital stats for term
            leader_board[term] = VitalStats{
                new_n,
                new_mean,
                new_sd,
            }
        } ()
    }

    select {}
}

func TermScores(c chan TermScore, n int) {
    words := [...]string{
        "magniloquent",
        "hetorodoxia",
        "fulgurant",
        "limpid",
        "sod",
    }

    for i := 0; i < n; i++ {
        word := words[rand.Intn(len(words))]
        score := rand.Float64()
        c <- TermScore{word, score}
    }
}