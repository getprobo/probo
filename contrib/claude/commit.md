# Commit Messages

Follow the [seven rules of a great Git commit message](https://cbea.ms/git-commit/):

1. Separate subject from body with a blank line
2. Limit the subject line to 50 characters
3. Capitalize the subject line
4. Do not end the subject line with a period
5. Use the imperative mood in the subject line
6. Wrap the body at 72 characters
7. Use the body to explain *what* and *why* vs. *how*

The subject line should complete the sentence: "If applied, this commit will *your subject line here*".

```
Add vendor assessment agent for third-party reviews

The existing changelog generator only covers internal changes.
This introduces a dedicated agent that evaluates third-party
vendors against our compliance criteria, producing a structured
risk report.
```

Not every commit needs a body -- a single line is fine when the change is self-explanatory:

```
Fix typo in vendor assessment prompt
```
