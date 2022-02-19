package store

/*
Store репозитории данных
*/
type Store interface {
	Friends() FriendsRepository // интерфейс для друзей
	User() UserRepository       // Интерфейс для друзей
}
