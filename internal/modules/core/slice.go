package core

func Map[TSource any, TResult any](source []TSource, m func(TSource) TResult) []TResult {
	results := make([]TResult, 0, len(source))
	for _, s := range source {
		results = append(results, m(s))
	}
	return results
}
