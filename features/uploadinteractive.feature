Feature: Interactives API (Get interactive)

    Scenario: POST an invalid interactive
        When I POST "/v1/interactives"
        """
            {
            }
        """
        Then the HTTP status code should be "400"