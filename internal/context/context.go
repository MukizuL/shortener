package context

type ContextKey string

// UserIDContextKey used as key for storing and fetching value from context.
const UserIDContextKey = ContextKey("userID")
