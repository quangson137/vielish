import { api } from "./api";

export interface Topic {
  id: string;
  name: string;
  name_vi: string;
  description: string;
  level: string;
}

export interface Word {
  id: string;
  word: string;
  ipa_phonetic: string;
  part_of_speech: string;
  vi_meaning: string;
  en_definition: string;
  example_sentence: string;
  example_vi_translation: string;
  audio_url: string;
  image_url: string;
  level: string;
  topic_id: string;
}

export interface QuizQuestion {
  word_id: string;
  word: string;
  options: string[];
}

export interface QuizResult {
  score: number;
  total: number;
  results: { word_id: string; correct: boolean; correct_answer: string }[];
}

export async function fetchTopics(level?: string): Promise<Topic[]> {
  const query = level ? `?level=${level}` : "";
  const res = await api.request(`/api/topics${query}`);
  if (!res.ok) throw new Error("Failed to fetch topics");
  return res.json();
}

export async function fetchTopicWords(topicId: string): Promise<Word[]> {
  const res = await api.request(`/api/topics/${topicId}/words`);
  if (!res.ok) throw new Error("Failed to fetch words");
  return res.json();
}

export async function fetchWord(id: string): Promise<Word> {
  const res = await api.request(`/api/words/${id}`);
  if (!res.ok) throw new Error("Failed to fetch word");
  return res.json();
}

export async function fetchDueReviews(): Promise<Word[]> {
  const res = await api.request("/api/review/due");
  if (!res.ok) throw new Error("Failed to fetch due reviews");
  return res.json();
}

export async function submitReview(
  wordId: string,
  quality: 1 | 3 | 5
): Promise<void> {
  const res = await api.request(`/api/review/${wordId}`, {
    method: "POST",
    body: JSON.stringify({ quality }),
  });
  if (!res.ok) throw new Error("Failed to submit review");
}

export async function fetchQuiz(
  topicId: string
): Promise<{ questions: QuizQuestion[] }> {
  const res = await api.request(`/api/quiz/${topicId}`);
  if (!res.ok) throw new Error("Failed to fetch quiz");
  return res.json();
}

export async function submitQuiz(
  topicId: string,
  answers: { word_id: string; answer: string }[]
): Promise<QuizResult> {
  const res = await api.request(`/api/quiz/${topicId}`, {
    method: "POST",
    body: JSON.stringify({ answers }),
  });
  if (!res.ok) throw new Error("Failed to submit quiz");
  return res.json();
}

export interface UserStats {
  streak: number;
  total_learned: number;
  due_today: number;
}

export async function fetchStats(): Promise<UserStats> {
  const res = await api.request("/api/stats");
  if (!res.ok) throw new Error("Failed to fetch stats");
  return res.json();
}
