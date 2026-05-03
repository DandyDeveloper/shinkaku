package db

import "fmt"

type starterGrammarPoint struct {
	jlptLevel string
	pattern   string
	meaning   string
	exampleJP string
	exampleEN string
	notes     string
}

var starterGrammarPoints = []starterGrammarPoint{
	{jlptLevel: "N5", pattern: "〜です", meaning: "to be; polite copula", exampleJP: "わたしはアメリカ人です。", exampleEN: "I am American.", notes: "Basic polite sentence ending."},
	{jlptLevel: "N5", pattern: "〜ます", meaning: "polite present or future verb ending", exampleJP: "毎日日本語を勉強します。", exampleEN: "I study Japanese every day.", notes: "Use with the verb stem in polite speech."},
	{jlptLevel: "N5", pattern: "〜ません", meaning: "polite negative verb ending", exampleJP: "今日はコーヒーを飲みません。", exampleEN: "I will not drink coffee today.", notes: "Polite negative form of 〜ます."},
	{jlptLevel: "N5", pattern: "〜ました", meaning: "polite past tense", exampleJP: "昨日、映画を見ました。", exampleEN: "I watched a movie yesterday.", notes: "Polite past form of verbs."},
	{jlptLevel: "N5", pattern: "〜たい", meaning: "want to do", exampleJP: "日本へ行きたいです。", exampleEN: "I want to go to Japan.", notes: "Attach to the verb stem."},
	{jlptLevel: "N5", pattern: "〜てください", meaning: "please do", exampleJP: "ここに名前を書いてください。", exampleEN: "Please write your name here.", notes: "Common polite request pattern."},
	{jlptLevel: "N5", pattern: "〜てもいい", meaning: "may do; it is okay to do", exampleJP: "ここに座ってもいいですか。", exampleEN: "May I sit here?", notes: "Permission pattern."},
	{jlptLevel: "N5", pattern: "〜てはいけない", meaning: "must not do", exampleJP: "ここで写真を撮ってはいけません。", exampleEN: "You must not take photos here.", notes: "Prohibition pattern."},
	{jlptLevel: "N5", pattern: "〜から", meaning: "because", exampleJP: "寒いですから、窓を閉めます。", exampleEN: "Because it is cold, I will close the window.", notes: "Reason or cause."},
	{jlptLevel: "N5", pattern: "〜ので", meaning: "because; since", exampleJP: "静かなので、このカフェが好きです。", exampleEN: "Because it is quiet, I like this cafe.", notes: "Often softer than 〜から."},
	{jlptLevel: "N4", pattern: "〜と思う", meaning: "think that", exampleJP: "あしたは雨が降ると思います。", exampleEN: "I think it will rain tomorrow.", notes: "Use to express opinions or guesses."},
	{jlptLevel: "N4", pattern: "〜つもりだ", meaning: "intend to do", exampleJP: "来年、日本に留学するつもりです。", exampleEN: "I intend to study abroad in Japan next year.", notes: "Shows a plan or intention."},
	{jlptLevel: "N4", pattern: "〜ながら", meaning: "while doing", exampleJP: "音楽を聞きながら勉強します。", exampleEN: "I study while listening to music.", notes: "Two simultaneous actions."},
	{jlptLevel: "N4", pattern: "〜ようになる", meaning: "come to be able to; come to do", exampleJP: "日本語の新聞が読めるようになりました。", exampleEN: "I have become able to read Japanese newspapers.", notes: "Describes change over time."},
	{jlptLevel: "N4", pattern: "〜なければならない", meaning: "must do", exampleJP: "明日までにレポートを出さなければなりません。", exampleEN: "I must submit the report by tomorrow.", notes: "Formal obligation pattern."},
	{jlptLevel: "N4", pattern: "〜かもしれない", meaning: "might; may", exampleJP: "彼はもう帰ったかもしれません。", exampleEN: "He might have already gone home.", notes: "Expresses uncertainty."},
	{jlptLevel: "N4", pattern: "〜たり〜たりする", meaning: "do things like A and B", exampleJP: "週末は映画を見たり、本を読んだりします。", exampleEN: "On weekends I do things like watch movies and read books.", notes: "Non-exhaustive list of actions."},
	{jlptLevel: "N4", pattern: "〜そうだ", meaning: "looks like; seems", exampleJP: "このケーキはおいしそうです。", exampleEN: "This cake looks delicious.", notes: "Judgment based on appearance."},
	{jlptLevel: "N3", pattern: "〜ことにする", meaning: "decide to do", exampleJP: "健康のために毎朝走ることにしました。", exampleEN: "I decided to run every morning for my health.", notes: "Speaker's decision."},
	{jlptLevel: "N3", pattern: "〜ようにする", meaning: "make an effort to; try to", exampleJP: "毎日漢字を復習するようにしています。", exampleEN: "I make an effort to review kanji every day.", notes: "Habitual effort or intention."},
	{jlptLevel: "N3", pattern: "〜はずだ", meaning: "should; expected to", exampleJP: "田中さんはもう駅に着いたはずです。", exampleEN: "Tanaka should have arrived at the station by now.", notes: "Expectation based on reasoning."},
	{jlptLevel: "N3", pattern: "〜わけではない", meaning: "it does not mean that; not necessarily", exampleJP: "高いものが全部いいわけではありません。", exampleEN: "It is not the case that all expensive things are good.", notes: "Partial negation."},
	{jlptLevel: "N3", pattern: "〜てしまう", meaning: "finish completely; do accidentally", exampleJP: "大事なメールを消してしまいました。", exampleEN: "I accidentally deleted an important email.", notes: "Often implies regret when accidental."},
	{jlptLevel: "N3", pattern: "〜おかげで", meaning: "thanks to", exampleJP: "先生のおかげで、試験に合格できました。", exampleEN: "Thanks to my teacher, I was able to pass the exam.", notes: "Used for positive causes."},
	{jlptLevel: "N2", pattern: "〜わけにはいかない", meaning: "cannot afford to; cannot do", exampleJP: "明日は大事な会議があるので、休むわけにはいきません。", exampleEN: "I have an important meeting tomorrow, so I cannot take the day off.", notes: "Social or situational impossibility."},
	{jlptLevel: "N2", pattern: "〜に違いない", meaning: "must be; there is no doubt", exampleJP: "あのレストランは毎日こんでいるから、おいしいに違いない。", exampleEN: "That restaurant is crowded every day, so it must be good.", notes: "Strong conviction based on evidence."},
	{jlptLevel: "N2", pattern: "〜ことになっている", meaning: "it is مقرر; it has been decided that", exampleJP: "この会社では九時までに出社することになっています。", exampleEN: "At this company, it is required that you arrive by nine.", notes: "Rule, plan, or arrangement."},
	{jlptLevel: "N2", pattern: "〜ものの", meaning: "although; but", exampleJP: "日本語を三年勉強したものの、まだ自信がありません。", exampleEN: "Although I have studied Japanese for three years, I still lack confidence.", notes: "Written or formal contrast."},
	{jlptLevel: "N1", pattern: "〜ざるを得ない", meaning: "cannot avoid doing; have no choice but to", exampleJP: "電車が止まったので、タクシーを使わざるを得ませんでした。", exampleEN: "Because the trains stopped, I had no choice but to take a taxi.", notes: "Formal expression of unavoidable action."},
	{jlptLevel: "N1", pattern: "〜にほかならない", meaning: "nothing but; none other than", exampleJP: "今回の成功はチーム全員の努力の結果にほかなりません。", exampleEN: "This success is nothing other than the result of the whole team's effort.", notes: "Emphatic formal identification."},
	{jlptLevel: "N1", pattern: "〜かたわら", meaning: "while; besides", exampleJP: "彼は会社で働くかたわら、小説も書いています。", exampleEN: "While working at a company, he also writes novels.", notes: "Formal pattern for parallel long-term activity."},
}

func (db *DB) seedStarterContent() error {
	var grammarCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM grammar_points`).Scan(&grammarCount); err != nil {
		return fmt.Errorf("count grammar points: %w", err)
	}
	if grammarCount > 0 {
		return nil
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin seed tx: %w", err)
	}
	defer tx.Rollback()

	insertGrammar, err := tx.Prepare(`INSERT INTO grammar_points (jlpt_level, pattern, meaning, example_jp, example_en, notes, source)
		VALUES (?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("prepare grammar insert: %w", err)
	}
	defer insertGrammar.Close()

	insertReviewCard, err := tx.Prepare(`INSERT INTO review_cards (grammar_point_id, interval, repetitions, e_factor, due_date)
		VALUES (?, 1, 0, 2.5, datetime('now'))`)
	if err != nil {
		return fmt.Errorf("prepare review card insert: %w", err)
	}
	defer insertReviewCard.Close()

	for _, gp := range starterGrammarPoints {
		res, err := insertGrammar.Exec(gp.jlptLevel, gp.pattern, gp.meaning, gp.exampleJP, gp.exampleEN, gp.notes, "starter")
		if err != nil {
			return fmt.Errorf("insert starter grammar %q: %w", gp.pattern, err)
		}
		grammarPointID, err := res.LastInsertId()
		if err != nil {
			return fmt.Errorf("starter grammar last insert id %q: %w", gp.pattern, err)
		}
		if _, err := insertReviewCard.Exec(grammarPointID); err != nil {
			return fmt.Errorf("insert review card for %q: %w", gp.pattern, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit starter seed: %w", err)
	}
	return nil
}