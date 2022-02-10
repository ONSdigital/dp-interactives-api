Feature: Interactives API (Get interactive)

    Scenario: POST an invalid interactive
        When I POST "/interactives"
        """
            {
            }
        """
        Then the HTTP status code should be "400"