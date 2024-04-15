# Trading Card Design Guidelines

Welcome to the Trading Card Design Guidelines for developers. This document 
provides essential guidelines for creating and managing trading cards within 
the Tony platform, ensuring consistency, user engagement, and seamless economic 
integration.

## 1. Overview

Trading cards in the Tony ecosystem can serve various purposes, including game 
elements, achievement rewards, or collectibles. This guide will help you 
nderstand how to effectively create and manage these cards.

## 2. Card Attributes

Each card should have the following attributes defined:

- **Name**: A unique identifier (e.g., `snailrace_achievement_first_win`). Keep 
        names concise and descriptive.
- **Title**: A short, descriptive title for the card (e.g., "Champion Snail").
- **Description**: A detailed description of the card, explaining its 
        significance or use or acquisition
- **Application**: The ID of the application the card belongs to.
- **Rarity**: Define the rarity (`common`, `uncommon`, `rare`, `epic`, 
        `legendary`).
- **Usable**: Boolean indicating if the card can be used in games or other 
        applications.
- **Tradable**: Boolean indicating if the card can be traded among users.
- **Unbreakable**: Boolean indicating if the card is immune to usage limits 
        (only applicable if Usable is true).
- **Max Usage**: The maximum number of times the card can be used (relevant 
        only if Usable is true).
- **Current Usage**: Tracks how many times the card has been used.
- **SVG**: An *Optional* graphical representation of the card in SVG format 
        with a resolution of `750x1050`.

> **Note:** Graphics are currently not viewable on the server until we have more
> cards to showcase.

## 3. Card Creation and Registration

### Creation Process

- Ensure that the card name is unique across the Tony platform.
- Provide all required attributes as per the schema defined above. Incomplete 
  cards will not be registered.

### Registration

- Use the `RegisterCard` API to submit your card for registration.
- Handle errors such as duplicate names or missing information gracefully.

## 4. Economic Interactions

- Cards can be marked as tradable or non-tradable.
- Tradable cards can be bought and sold using using funds from Tony's wallet.
- Ensure that the economic activities related to cards do not disrupt the game 
  or application balance.

## 5. Guidelines for Card Usage

- Define clear rules for how cards can be used within your applications.
- If cards have a usage limit, ensure this is clearly communicated to users.
- Implement checks to ensure cards are not used beyond their maximum usage. 
  Otherwise, the registry will delete the card link on usage of 0 and will 
  return `ErrCardNotFound` if attempted to be used.

## 6. Updates and Maintenance

- Update the card attributes and functionalities based on user feedback and 
  platform updates. Cards are linked via their unique name, so you can change
  details about without worrying about users losing their cards.
- Monitor and analyze the usage of cards to ensure they are meeting their 
  intended purposes effectively.

By adhering to these guidelines, developers can create engaging and functional 
trading cards that enhance user experience and contribute to a vibrant 
community ecosystem.
