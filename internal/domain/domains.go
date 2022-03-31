// Package domains holds the bussines logic
// The domain layer contains the components of the application that describe the business foundations:
// 	* Domain Entities & Value Objects
// 		You can think of these as the data structures of the business domain
// 		Entities are uniquely identifiable structs
// 		Value Objects are structs that are not uniquely identifiable, and they describe the characteristics of an Object.
//  * Domain Services
// 		They provide domain functionality using entities and objects i.e., data repository
//
// The domain layer does not have any dependencies on other layers. It is also only used by the application layer.
package domain
