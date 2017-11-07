package main

import (
	"bufio"
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

var quizFile = flag.String("f", "problems.csv", "file with the quiz in it")
var timeLimit = flag.String("t", "30", "time limit in seconds")
var randomize = flag.Bool("r", false, "randomize the quiz")

func main() {
	flag.Parse()

	q, err := getQuiz()
	if err != nil {
		panic(err)
	}

	if *randomize {
		randomizeQuiz(q)
	}

	limit, err := strconv.Atoi(*timeLimit)
	if err != nil {
		panic(err)
	}

	r := bufio.NewReader(os.Stdin)
	fmt.Print("Let's take a quiz! Hit enter when you're ready to start ")
	r.ReadString('\n')

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(limit)*time.Second)
	defer cancel()

	startQuiz(ctx, q)
	fmt.Printf("The quiz is complete, you answered %d out of %d problems correctly\n", q.numCorrect, len(q.questions))
}

func startQuiz(ctx context.Context, q *quiz) {
	fmt.Printf("You have %s seconds\n", *timeLimit)

	questionsDone := make(chan bool, 1)

	go askQuestions(q, questionsDone)

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Ran out of time!")
			return
		case <-questionsDone:
			fmt.Println("Finished all ?'s")
			return
		}
	}
}

func askQuestions(q *quiz, questionsDone chan bool) {
	for _, question := range q.questions {
		correct, err := ask(question)
		if err != nil {
			panic(err)
		}

		if correct {
			q.numCorrect++
		}
	}

	questionsDone <- true
}

func ask(q *question) (bool, error) {
	r := bufio.NewReader(os.Stdin)

	fmt.Printf("What is %s? ", q.operands)
	ans, err := r.ReadString('\n')
	if err != nil {
		return false, err
	}

	ans = strings.Replace(ans, "\n", "", -1)
	if ans == "" {
		return false, nil
	}

	return ans == q.answer, nil
}

type quiz struct {
	questions  []*question
	numCorrect int
}

type question struct {
	operands string
	answer   string
}

const (
	QUESTION_IDX = iota
	ANSWER_IDX
)

func getQuiz() (*quiz, error) {
	fp, err := os.Open(*quizFile)
	if err != nil {
		return nil, err
	}

	records, err := csv.NewReader(fp).ReadAll()
	if err != nil {
		return nil, err
	}

	q := &quiz{
		questions: make([]*question, 0),
	}

	for _, rec := range records {
		q.questions = append(q.questions, &question{
			operands: rec[QUESTION_IDX],
			answer:   rec[ANSWER_IDX],
		})
	}

	return q, nil
}

func randomizeQuiz(q *quiz) {
	rand.Seed(time.Now().Unix())
	list := rand.Perm(len(q.questions))

	randQuestions := make([]*question, len(q.questions))
	for i, v := range list {
		randQuestions[v] = q.questions[i]
	}

	q.questions = randQuestions
}
