'use client';

import React, { memo, useMemo } from 'react';

interface DiffViewerProps {
  diff: string;
  fileName?: string;
}

interface DiffLine {
  type: 'added' | 'removed' | 'context' | 'header';
  content: string;
  lineNumber?: number;
}

function parseDiff(diff: string): DiffLine[] {
  const lines = diff.split('\n');
  const result: DiffLine[] = [];
  let oldLineNumber = 0;
  let newLineNumber = 0;

  for (const line of lines) {
    if (line.startsWith('@@')) {
      // Parse hunk header
      const match = line.match(/@@ -(\d+),?\d* \+(\d+),?\d* @@/);
      if (match) {
        oldLineNumber = parseInt(match[1], 10);
        newLineNumber = parseInt(match[2], 10);
      }
      result.push({ type: 'header', content: line });
    } else if (line.startsWith('+') && !line.startsWith('+++')) {
      result.push({ type: 'added', content: line.slice(1), lineNumber: newLineNumber });
      newLineNumber++;
    } else if (line.startsWith('-') && !line.startsWith('---')) {
      result.push({ type: 'removed', content: line.slice(1), lineNumber: oldLineNumber });
      oldLineNumber++;
    } else if (line.startsWith(' ')) {
      result.push({ type: 'context', content: line.slice(1), lineNumber: newLineNumber });
      oldLineNumber++;
      newLineNumber++;
    } else if (line.startsWith('diff ') || line.startsWith('index ') || line.startsWith('---') || line.startsWith('+++')) {
      result.push({ type: 'header', content: line });
    }
  }

  return result;
}

const lineTypeStyles: Record<DiffLine['type'], string> = {
  added: 'bg-green-900/30 text-green-300',
  removed: 'bg-red-900/30 text-red-300',
  context: 'text-gray-400',
  header: 'bg-gray-700 text-gray-300 font-mono',
};

const lineTypeIndicators: Record<DiffLine['type'], string> = {
  added: '+',
  removed: '-',
  context: ' ',
  header: '',
};

export const DiffViewer = memo(function DiffViewer({ diff, fileName }: DiffViewerProps) {
  // Memoize expensive diff parsing
  const parsedLines = useMemo(() => parseDiff(diff), [diff]);

  if (!diff || parsedLines.length === 0) {
    return (
      <div className="text-center py-8 text-gray-400 bg-gray-800 rounded-lg">
        No diff content available
      </div>
    );
  }

  return (
    <div className="rounded-lg overflow-hidden border border-gray-700">
      {fileName && (
        <div className="bg-gray-800 px-4 py-2 border-b border-gray-700">
          <span className="text-sm font-mono text-gray-300">{fileName}</span>
        </div>
      )}
      <div className="overflow-x-auto">
        <pre className="text-sm leading-6">
          {parsedLines.map((line, index) => (
            <div
              key={index}
              className={`px-4 py-0.5 font-mono ${lineTypeStyles[line.type]}`}
            >
              <span className="inline-block w-6 text-gray-500 select-none">
                {lineTypeIndicators[line.type]}
              </span>
              {line.lineNumber !== undefined && (
                <span className="inline-block w-12 text-gray-500 select-none text-right pr-4">
                  {line.lineNumber}
                </span>
              )}
              <span className="whitespace-pre">{line.content}</span>
            </div>
          ))}
        </pre>
      </div>
    </div>
  );
});
