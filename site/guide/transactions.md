# Transactions

Sometimes several writes must happen together or not at all: create an order and reserve its items, or move a task between lists and update both counters. A transaction groups them. If anything fails, none of them stays.

::: info What you'll learn
- How to run code in a transaction
- How the transaction reaches every repository without changing them
- What happens with errors, panics and nested transactions
:::

## Run a transaction

Embed `database.Transactional` in a component, and wrap the work in `s.Transaction`:

```go
type Service struct {
	database.Transactional // gives the service s.Transaction
	Lists *Repository      `inject:""`
	Tasks *task.Service    `inject:""`
}

func (s *Service) Archive(ctx context.Context, id string) error {
	return s.Transaction(ctx, func(ctx context.Context) error {
		if _, err := s.Lists.Update(ctx, id, database.Args{"archived": true}); err != nil {
			return err
		}
		return s.Tasks.CompleteAll(ctx, id)
	})
}
```

- **Commit** when the function returns `nil`.
- **Rollback** when it returns an error, or when it panics. The panic goes on after the rollback.

## The transaction travels in `ctx`

The function receives a new `ctx` that carries the transaction. Every query that runs with that `ctx` joins the transaction: the repository's operations, `r.Prepare` and the services of other preás. The repositories themselves don't change; they don't know whether they are inside a transaction.

::: warning Use the `ctx` the function receives
```go
s.Transaction(ctx, func(txCtx context.Context) error {
	s.Lists.Update(ctx, id, args)   // ✗ the outer ctx: runs outside the transaction
	s.Lists.Update(txCtx, id, args) // ✓
})
```
Naming the parameter `ctx`, as in the first example, hides the outer one and avoids the mistake.
:::

## Nested transactions

A transaction inside another one opens a **savepoint**. If the inner one fails, only its work is undone, and the outer one decides what to do with the error: return it, which undoes everything, or handle it and go on.

```go
return s.Transaction(ctx, func(ctx context.Context) error {
	if err := s.Orders.Create(ctx, order); err != nil {
		return err
	}
	// Nice to have: a failure here undoes only the coupon.
	if err := s.Coupons.Apply(ctx, order); err != nil { // runs its own Transaction
		logger.Info("coupon not applied", "error", err)
	}
	return nil
})
```

## Things to keep in mind

- **A transaction is one connection.** Don't use its `ctx` from several goroutines at the same time.
- **Keep it short.** A transaction holds locks until it ends. Don't call slow external services inside one.
- **Raw pool access stays outside.** Calling `r.DB.Query` directly skips the transaction. Use `r.Prepare` or the repository's operations.

## Next steps

- [HTTP security](./http-security)
- [Testing](./testing)
