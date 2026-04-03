"use client";

import { useState, useEffect } from "react";
import { Word } from "@/lib/vocab-api";

interface FlashcardProps {
  word: Word;
  onRate: (quality: 1 | 3 | 5) => void;
}

export default function Flashcard({ word, onRate }: FlashcardProps) {
  const [flipped, setFlipped] = useState(false);
  const [hasTTS, setHasTTS] = useState(false);

  useEffect(() => {
    setHasTTS(typeof window !== "undefined" && "speechSynthesis" in window);
  }, []);

  useEffect(() => {
    setFlipped(false);
  }, [word]);

  const speak = () => {
    const utterance = new SpeechSynthesisUtterance(word.word);
    utterance.lang = "en-US";
    window.speechSynthesis.speak(utterance);
  };

  return (
    <div className="max-w-md mx-auto">
      <div
        onClick={() => setFlipped(!flipped)}
        className="relative border border-warm-border bg-warm-surface rounded-lg p-8 min-h-[250px] flex flex-col items-center justify-center cursor-pointer hover:shadow-md transition-shadow"
      >
        {hasTTS && (
          <button
            onClick={(e) => {
              e.stopPropagation();
              speak();
            }}
            className="absolute top-4 right-4 text-warm-subtle hover:text-warm-accent transition-colors text-xl"
            aria-label="Phát âm"
          >
            🔊
          </button>
        )}

        {!flipped ? (
          <>
            <p className="text-3xl font-bold mb-2 text-warm-text">{word.word}</p>
            {word.ipa_phonetic && (
              <p className="text-warm-subtle">{word.ipa_phonetic}</p>
            )}
            {word.part_of_speech && (
              <p className="text-warm-subtle text-sm mt-1">
                {word.part_of_speech}
              </p>
            )}
            <p className="text-warm-subtle text-sm mt-4">Nhấn để lật thẻ</p>
          </>
        ) : (
          <>
            <p className="text-2xl font-semibold text-warm-accent mb-3">
              {word.vi_meaning}
            </p>
            {word.en_definition && (
              <p className="text-warm-muted text-sm mb-2">
                {word.en_definition}
              </p>
            )}
            {word.example_sentence && (
              <div className="mt-3 text-sm text-warm-subtle">
                <p className="italic">{word.example_sentence}</p>
                {word.example_vi_translation && (
                  <p className="text-warm-subtle mt-1">
                    {word.example_vi_translation}
                  </p>
                )}
              </div>
            )}
          </>
        )}
      </div>

      {flipped && (
        <div className="flex justify-center gap-3 mt-4">
          <button
            onClick={() => onRate(1)}
            className="px-6 py-2 bg-red-500 text-white rounded hover:bg-red-600"
          >
            Khó
          </button>
          <button
            onClick={() => onRate(3)}
            className="px-6 py-2 bg-yellow-500 text-white rounded hover:bg-yellow-600"
          >
            Ổn
          </button>
          <button
            onClick={() => onRate(5)}
            className="px-6 py-2 bg-green-500 text-white rounded hover:bg-green-600"
          >
            Dễ
          </button>
        </div>
      )}
    </div>
  );
}
