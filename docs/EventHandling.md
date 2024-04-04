# Event Handling

Some applications/commands you wish to build may utilise Discords Interactions
and Modal features. When these interaction events are made they will need to be
routed back to the right application command. Every interaction requires a 
`custom_id` value, for the purposes of **Tony** these need to be specifically
formatted in the following way:

```
command{.subcommand}:value{:value}
```

An example usecase for this can be placing a bet of $64, the `custom_id` may be
`bet.place:64` to route back to the placing a bet application. Some applications
might require additional data like a reference to a particular thing to place a 
bet on, the picks and how much to place:

```
bet.place:melbourne_cup_race_4:3:4:2:30


Breakdown:

    Route   = bet.place
    Race Id = melbourne_cup_race_4
    Pick 1  = 3
    Pick 2  = 4
    Pick 3  = 2
    Price   = 30
```