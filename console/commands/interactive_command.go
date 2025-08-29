package commands

import (
	"fmt"
)

type InteractiveCommand struct {
	BaseCommand
}

func (c *InteractiveCommand) GetSignature() string {
	return "interactive:demo"
}

func (c *InteractiveCommand) GetDescription() string {
	return "Demonstrate interactive console input methods"
}

func (c *InteractiveCommand) Execute(args []string) error {
	fmt.Println("🎯 Interactive Console Demo")
	fmt.Println("This command demonstrates various input methods:")
	fmt.Println()

	// 1. Simple text input
	name := c.AskText("What's your name?", "Anonymous")
	fmt.Printf("Hello, %s! 👋\n\n", name)

	// 2. Required input (won't accept empty)
	email := c.AskRequired("Enter your email:")
	fmt.Printf("Email: %s ✅\n\n", email)

	// 3. Yes/No confirmation
	confirmed := c.AskConfirmation("Do you want to continue?", true)
	if !confirmed {
		fmt.Println("Operation cancelled! 👋")
		return nil
	}
	fmt.Println("Great! Continuing... 🚀")

	// 4. Multiple choice
	choices := []string{"Go", "Python", "JavaScript", "Rust"}
	language := c.AskChoice("What's your favorite programming language?", choices, 0)
	fmt.Printf("Nice choice! %s is awesome! 🎉\n\n", language)

	// 5. Numeric input
	age := c.AskNumber("How old are you?", 25)
	fmt.Printf("Age: %d years old\n\n", age)

	// 6. Password input (hidden)
	password := c.AskPassword("Enter a password:")
	fmt.Printf("Password length: %d characters (hidden for security) 🔒\n\n", len(password))

	// 7. List input (comma-separated)
	hobbies := c.AskList("What are your hobbies? (comma-separated)", []string{"coding"})
	fmt.Printf("Hobbies: %v 🎨\n\n", hobbies)

	fmt.Println("✅ Interactive demo completed!")
	fmt.Printf("Summary: %s (%s), %d years old, loves %s\n", name, email, age, language)

	return nil
}
