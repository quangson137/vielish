"use client";

import { useState } from "react";
import { Word } from "@/lib/vocab-api";

interface FlashcardProps {
  word: Word;
  onRate: (quality: 1 | 3 | 5) => void;
}

export default function Flashcard({ word, onRate }: FlashcardProps) {
  const [flipped, setFlipped] = useState(false);

  return (
    <div className="max-w-md mx-auto">
      <div
        onClick={() => setFlipped(!flipped)}
        className="border rounded-lg p-8 min-h-[250px] flex flex-col items-center justify-center cursor-pointer hover:shadow-md transition-shadow"
      >
        {!flipped ? (
          <>
            <p className="text-3xl font-bold mb-2">{word.word}</p>
            {word.ipa_phonetic && (
              <p className="text-gray-500">{word.ipa_phonetic}</p>
            )}
            {word.part_of_speech && (
              <p className="text-gray-400 text-sm mt-1">
                {word.part_of_speech}
              </p>
            )}
            <p className="text-gray-400 text-sm mt-4">Nhấn để lật thẻ</p>
          </>
        ) : (
          <>
            <p className="text-2xl font-semibold text-blue-700 mb-3">
              {word.vi_meaning}
            </p>
            {word.en_definition && (
              <p className="text-gray-600 text-sm mb-2">
                {word.en_definition}
              </p>
            )}
            {word.example_sentence && (
              <div className="mt-3 text-sm text-gray-500">
                <p className="italic">{word.example_sentence}</p>
                {word.example_vi_translation && (
                  <p className="text-gray-400 mt-1">
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
