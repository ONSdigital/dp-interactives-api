Feature: Interactives API (Delete interactive)

    Scenario: Delete failed if interactive not in DB
        When I DELETE "/v1/interactives/0d77a889-abb2-4432-ad22-9c23cf7ee796"
        Then the HTTP status code should be "404"

    Scenario: Successful delete
    Given I have these interactives:
            """
            [
                {
                    "_id": "0d77a889-abb2-4432-ad22-9c23cf7ee796",
                    "active": true,
                    "file_name": "kqA7qPo1GeOJeff69lByWLbPiZM=/docker-vernemq-master.zip",
                    "last_updated": {
                        "$date": "2022-02-08T19:04:52.891Z"
                    },
                    "metadata": "{\"metadata1\":\"XXX\",\"metadata2\":\"YYY\",\"metadata3\":\"ZZZ\"}",
                    "sha": "kqA7qPo1GeOJeff69lByWLbPiZM=",
                    "state": "ArchiveUploaded"
                }
            ]
            """
    When I DELETE "/v1/interactives/0d77a889-abb2-4432-ad22-9c23cf7ee796"
    Then the HTTP status code should be "200"