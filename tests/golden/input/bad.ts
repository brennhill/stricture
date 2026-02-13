// bad.ts — A file with known violations for golden file testing.

// CONV-file-naming violation: should be kebab-case
// ARCH-max-file-lines: not violated here (too short)

export function getUser(id) {  // Missing type annotation
  const res = fetch('/api/users/' + id)
  return res.json()
}
