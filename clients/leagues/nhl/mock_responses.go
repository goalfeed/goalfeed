package nhl

const ActiveGameScoreboard = `
// 20210104162225
// https://statsapi.web.nhl.com/api/v1/game/2020020001/feed/live

{
  "copyright": "NHL and the NHL Shield are registered trademarks of the National Hockey League. NHL and NHL team marks are the property of the NHL and its teams. © NHL 2021. All Rights Reserved.",
  "gamePk": 2020020001,
  "link": "/api/v1/game/2020020001/feed/live",
  "metaData": {
    "wait": 10,
    "timeStamp": "20210104_214258"
  },
  "gameData": {
    "game": {
      "pk": 2020020001,
      "season": "20202021",
      "type": "R"
    },
    "datetime": {
      "dateTime": "2021-01-13T22:30:00Z"
    },
    "status": {
      "abstractGameState": "Preview",
      "codedGameState": "1",
      "detailedState": "Live",
      "statusCode": "1",
      "startTimeTBD": false
    },
    "teams": {
      "away": {
        "id": 5,
        "name": "Pittsburgh Penguins",
        "link": "/api/v1/teams/5",
        "venue": {
          "id": 5034,
          "name": "PPG Paints Arena",
          "link": "/api/v1/venues/5034",
          "city": "Pittsburgh",
          "timeZone": {
            "id": "America/New_York",
            "offset": -5,
            "tz": "EST"
          }
        },
        "abbreviation": "PIT",
        "triCode": "PIT",
        "teamName": "Penguins",
        "locationName": "Pittsburgh",
        "firstYearOfPlay": "1967",
        "division": {
          "id": 25,
          "name": "East",
          "link": "/api/v1/divisions/25"
        },
        "conference": {
          "id": 6,
          "name": "Eastern",
          "link": "/api/v1/conferences/6"
        },
        "franchise": {
          "franchiseId": 17,
          "teamName": "Penguins",
          "link": "/api/v1/franchises/17"
        },
        "shortName": "Pittsburgh",
        "officialSiteUrl": "http://pittsburghpenguins.com/",
        "franchiseId": 17,
        "active": true
      },
      "home": {
        "id": 4,
        "name": "Philadelphia Flyers",
        "link": "/api/v1/teams/4",
        "venue": {
          "id": 5096,
          "name": "Wells Fargo Center",
          "link": "/api/v1/venues/5096",
          "city": "Philadelphia",
          "timeZone": {
            "id": "America/New_York",
            "offset": -5,
            "tz": "EST"
          }
        },
        "abbreviation": "PHI",
        "triCode": "PHI",
        "teamName": "Flyers",
        "locationName": "Philadelphia",
        "firstYearOfPlay": "1967",
        "division": {
          "id": 25,
          "name": "East",
          "link": "/api/v1/divisions/25"
        },
        "conference": {
          "id": 6,
          "name": "Eastern",
          "link": "/api/v1/conferences/6"
        },
        "franchise": {
          "franchiseId": 16,
          "teamName": "Flyers",
          "link": "/api/v1/franchises/16"
        },
        "shortName": "Philadelphia",
        "officialSiteUrl": "http://www.philadelphiaflyers.com/",
        "franchiseId": 16,
        "active": true
      }
    },
    "players": {
      
    },
    "venue": {
      "id": 5096,
      "name": "Wells Fargo Center",
      "link": "/api/v1/venues/5096"
    }
  },
  "liveData": {
    "plays": {
      "allPlays": [
        
      ],
      "scoringPlays": [
        
      ],
      "penaltyPlays": [
        
      ],
      "playsByPeriod": [
        
      ]
    },
    "linescore": {
      "currentPeriod": 0,
      "periods": [
        
      ],
      "shootoutInfo": {
        "away": {
          "scores": 0,
          "attempts": 0
        },
        "home": {
          "scores": 0,
          "attempts": 0
        }
      },
      "teams": {
        "home": {
          "team": {
            "id": 4,
            "name": "Philadelphia Flyers",
            "link": "/api/v1/teams/4",
            "abbreviation": "PHI",
            "triCode": "PHI"
          },
          "goals": 0,
          "shotsOnGoal": 0,
          "goaliePulled": false,
          "numSkaters": 0,
          "powerPlay": false
        },
        "away": {
          "team": {
            "id": 5,
            "name": "Pittsburgh Penguins",
            "link": "/api/v1/teams/5",
            "abbreviation": "PIT",
            "triCode": "PIT"
          },
          "goals": 0,
          "shotsOnGoal": 0,
          "goaliePulled": false,
          "numSkaters": 0,
          "powerPlay": false
        }
      },
      "powerPlayStrength": "Even",
      "hasShootout": false,
      "intermissionInfo": {
        "intermissionTimeRemaining": 0,
        "intermissionTimeElapsed": 0,
        "inIntermission": false
      }
    },
    "boxscore": {
      "teams": {
        "away": {
          "team": {
            "id": 5,
            "name": "Pittsburgh Penguins",
            "link": "/api/v1/teams/5",
            "abbreviation": "PIT",
            "triCode": "PIT"
          },
          "teamStats": {
            "teamSkaterStats": {
              "goals": 0,
              "pim": 0,
              "shots": 0,
              "powerPlayPercentage": "0.0",
              "powerPlayGoals": 0.0,
              "powerPlayOpportunities": 0.0,
              "faceOffWinPercentage": "0.0",
              "blocked": 0,
              "takeaways": 0,
              "giveaways": 0,
              "hits": 0
            }
          },
          "players": {
            
          },
          "goalies": [
            
          ],
          "skaters": [
            
          ],
          "onIce": [
            
          ],
          "onIcePlus": [
            
          ],
          "scratches": [
            
          ],
          "penaltyBox": [
            
          ],
          "coaches": [
            
          ]
        },
        "home": {
          "team": {
            "id": 4,
            "name": "Philadelphia Flyers",
            "link": "/api/v1/teams/4",
            "abbreviation": "PHI",
            "triCode": "PHI"
          },
          "teamStats": {
            "teamSkaterStats": {
              "goals": 0,
              "pim": 0,
              "shots": 0,
              "powerPlayPercentage": "0.0",
              "powerPlayGoals": 0.0,
              "powerPlayOpportunities": 0.0,
              "faceOffWinPercentage": "0.0",
              "blocked": 0,
              "takeaways": 0,
              "giveaways": 0,
              "hits": 0
            }
          },
          "players": {
            
          },
          "goalies": [
            
          ],
          "skaters": [
            
          ],
          "onIce": [
            
          ],
          "onIcePlus": [
            
          ],
          "scratches": [
            
          ],
          "penaltyBox": [
            
          ],
          "coaches": [
            
          ]
        }
      },
      "officials": [
        
      ]
    },
    "decisions": {
      
    }
  }
}
`

// UpcomingGamesSchedule JSON Response for the NHL schedule. Contains 1 active, 1 final and several upcoming games
const UpcomingGamesSchedule = `
{
  "copyright": "NHL and the NHL Shield are registered trademarks of the National Hockey League. NHL and NHL team marks are the property of the NHL and its teams. © NHL 2021. All Rights Reserved.",
  "totalItems": 5,
  "totalEvents": 0,
  "totalGames": 5,
  "totalMatches": 0,
  "wait": 10,
  "dates": [
    {
      "date": "2021-01-13",
      "totalItems": 5,
      "totalEvents": 0,
      "totalGames": 5,
      "totalMatches": 0,
      "games": [
        {
          "gamePk": 2020020001,
          "link": "/api/v1/game/2020020001/feed/live",
          "gameType": "R",
          "season": "20202021",
          "gameDate": "2021-01-13T22:30:00Z",
          "status": {
            "abstractGameState": "Live",
            "codedGameState": "1",
            "detailedState": "Live",
            "statusCode": "1",
            "startTimeTBD": false
          },
          "teams": {
            "away": {
              "leagueRecord": {
                "wins": 0,
                "losses": 0,
                "ot": 0,
                "type": "league"
              },
              "score": 0,
              "team": {
                "id": 5,
                "name": "Pittsburgh Penguins",
                "link": "/api/v1/teams/5"
              }
            },
            "home": {
              "leagueRecord": {
                "wins": 0,
                "losses": 0,
                "ot": 0,
                "type": "league"
              },
              "score": 0,
              "team": {
                "id": 4,
                "name": "Philadelphia Flyers",
                "link": "/api/v1/teams/4"
              }
            }
          },
          "venue": {
            "id": 5096,
            "name": "Wells Fargo Center",
            "link": "/api/v1/venues/5096"
          },
          "content": {
            "link": "/api/v1/game/2020020001/content"
          }
        },
        {
          "gamePk": 2020020003,
          "link": "/api/v1/game/2020020003/feed/live",
          "gameType": "R",
          "season": "20202021",
          "gameDate": "2021-01-14T00:00:00Z",
          "status": {
            "abstractGameState": "Preview",
            "codedGameState": "1",
            "detailedState": "Scheduled",
            "statusCode": "1",
            "startTimeTBD": false
          },
          "teams": {
            "away": {
              "leagueRecord": {
                "wins": 0,
                "losses": 0,
                "ot": 0,
                "type": "league"
              },
              "score": 0,
              "team": {
                "id": 8,
                "name": "Montréal Canadiens",
                "link": "/api/v1/teams/8"
              }
            },
            "home": {
              "leagueRecord": {
                "wins": 0,
                "losses": 0,
                "ot": 0,
                "type": "league"
              },
              "score": 0,
              "team": {
                "id": 10,
                "name": "Toronto Maple Leafs",
                "link": "/api/v1/teams/10"
              }
            }
          },
          "venue": {
            "name": "Scotiabank Arena",
            "link": "/api/v1/venues/null"
          },
          "content": {
            "link": "/api/v1/game/2020020003/content"
          }
        },
        {
          "gamePk": 2020020002,
          "link": "/api/v1/game/2020020002/feed/live",
          "gameType": "R",
          "season": "20202021",
          "gameDate": "2021-01-14T01:00:00Z",
          "status": {
            "abstractGameState": "Preview",
            "codedGameState": "1",
            "detailedState": "Scheduled",
            "statusCode": "1",
            "startTimeTBD": false
          },
          "teams": {
            "away": {
              "leagueRecord": {
                "wins": 0,
                "losses": 0,
                "ot": 0,
                "type": "league"
              },
              "score": 0,
              "team": {
                "id": 16,
                "name": "Chicago Blackhawks",
                "link": "/api/v1/teams/16"
              }
            },
            "home": {
              "leagueRecord": {
                "wins": 0,
                "losses": 0,
                "ot": 0,
                "type": "league"
              },
              "score": 0,
              "team": {
                "id": 14,
                "name": "Tampa Bay Lightning",
                "link": "/api/v1/teams/14"
              }
            }
          },
          "venue": {
            "id": 5017,
            "name": "Amalie Arena",
            "link": "/api/v1/venues/5017"
          },
          "content": {
            "link": "/api/v1/game/2020020002/content"
          }
        },
        {
          "gamePk": 2020020004,
          "link": "/api/v1/game/2020020004/feed/live",
          "gameType": "R",
          "season": "20202021",
          "gameDate": "2021-01-14T03:00:00Z",
          "status": {
            "abstractGameState": "Preview",
            "codedGameState": "1",
            "detailedState": "Scheduled",
            "statusCode": "1",
            "startTimeTBD": false
          },
          "teams": {
            "away": {
              "leagueRecord": {
                "wins": 0,
                "losses": 0,
                "ot": 0,
                "type": "league"
              },
              "score": 0,
              "team": {
                "id": 23,
                "name": "Vancouver Canucks",
                "link": "/api/v1/teams/23"
              }
            },
            "home": {
              "leagueRecord": {
                "wins": 0,
                "losses": 0,
                "ot": 0,
                "type": "league"
              },
              "score": 0,
              "team": {
                "id": 22,
                "name": "Edmonton Oilers",
                "link": "/api/v1/teams/22"
              }
            }
          },
          "venue": {
            "id": 5100,
            "name": "Rogers Place",
            "link": "/api/v1/venues/5100"
          },
          "content": {
            "link": "/api/v1/game/2020020004/content"
          }
        },
        {
          "gamePk": 2020020005,
          "link": "/api/v1/game/2020020005/feed/live",
          "gameType": "R",
          "season": "20202021",
          "gameDate": "2021-01-14T03:30:00Z",
          "status": {
            "abstractGameState": "Preview",
            "codedGameState": "1",
            "detailedState": "Scheduled",
            "statusCode": "1",
            "startTimeTBD": false
          },
          "teams": {
            "away": {
              "leagueRecord": {
                "wins": 0,
                "losses": 0,
                "ot": 0,
                "type": "league"
              },
              "score": 0,
              "team": {
                "id": 19,
                "name": "St. Louis Blues",
                "link": "/api/v1/teams/19"
              }
            },
            "home": {
              "leagueRecord": {
                "wins": 0,
                "losses": 0,
                "ot": 0,
                "type": "league"
              },
              "score": 0,
              "team": {
                "id": 21,
                "name": "Colorado Avalanche",
                "link": "/api/v1/teams/21"
              }
            }
          },
          "venue": {
            "id": 5064,
            "name": "Ball Arena",
            "link": "/api/v1/venues/5064"
          },
          "content": {
            "link": "/api/v1/game/2020020005/content"
          }
        }
      ],
      "events": [
        
      ],
      "matches": [
        
      ]
    }
  ]
}
`

const TeamResponseJson = `
{
  "copyright": "NHL and the NHL Shield are registered trademarks of the National Hockey League. NHL and NHL team marks are the property of the NHL and its teams. © NHL 2021. All Rights Reserved.",
  "teams": [
    {
      "id": 5,
      "name": "Pittsburgh Penguins",
      "link": "/api/v1/teams/5",
      "venue": {
        "id": 5034,
        "name": "PPG Paints Arena",
        "link": "/api/v1/venues/5034",
        "city": "Pittsburgh",
        "timeZone": {
          "id": "America/New_York",
          "offset": -5,
          "tz": "EST"
        }
      },
      "abbreviation": "PIT",
      "teamName": "Penguins",
      "locationName": "Pittsburgh",
      "firstYearOfPlay": "1967",
      "division": {
        "id": 25,
        "name": "MassMutual East",
        "link": "/api/v1/divisions/25"
      },
      "conference": {
        "id": 6,
        "name": "Eastern",
        "link": "/api/v1/conferences/6"
      },
      "franchise": {
        "franchiseId": 17,
        "teamName": "Penguins",
        "link": "/api/v1/franchises/17"
      },
      "shortName": "Pittsburgh",
      "officialSiteUrl": "http://pittsburghpenguins.com/",
      "franchiseId": 17,
      "active": true
    }
  ]
}
`
const DiffPatchResponseJson = `
[
  {
    "diff": [
      {
        "op": "replace",
        "path": "/metaData/timeStamp",
        "value": "20210122_040323"
      },
      {
        "op": "replace",
        "path": "/liveData/plays/allPlays/115/result/description",
        "value": "Tyler Toffoli (5) Wrist Shot, assists: Joel Armia (3)"
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/away/players/ID8474038/stats/skaterStats/timeOnIce",
        "value": "6:41"
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/away/players/ID8474038/stats/skaterStats/evenTimeOnIce",
        "value": "4:53"
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/away/players/ID8477476/stats/skaterStats/timeOnIce",
        "value": "5:57"
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/away/players/ID8477476/stats/skaterStats/evenTimeOnIce",
        "value": "3:45"
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/away/players/ID8474596/stats/goalieStats/timeOnIce",
        "value": "31:57"
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/away/players/ID8475279/stats/skaterStats/timeOnIce",
        "value": "10:53"
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/away/players/ID8475279/stats/skaterStats/evenTimeOnIce",
        "value": "6:15"
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/away/players/ID8478133/stats/skaterStats/timeOnIce",
        "value": "8:31"
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/away/players/ID8478133/stats/skaterStats/evenTimeOnIce",
        "value": "6:36"
      },
      {
        "op": "replace",
        "path": "/liveData/linescore/teams/away/goals",
        "value": "9:10"
      },
      {
        "op": "replace",
        "path": "/liveData/linescore/teams/home/goals",
        "value": 2
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/home/players/ID8474849/stats/skaterStats/timeOnIce",
        "value": "5:07"
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/home/players/ID8474849/stats/skaterStats/evenTimeOnIce",
        "value": "2:55"
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/home/players/ID8476468/stats/skaterStats/timeOnIce",
        "value": "10:06"
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/home/players/ID8476468/stats/skaterStats/evenTimeOnIce",
        "value": "5:08"
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/home/players/ID8477967/stats/goalieStats/timeOnIce",
        "value": "29:18"
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/home/players/ID8478444/stats/skaterStats/timeOnIce",
        "value": "10:42"
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/home/players/ID8478444/stats/skaterStats/evenTimeOnIce",
        "value": "5:55"
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/home/players/ID8480800/stats/skaterStats/timeOnIce",
        "value": "15:11"
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/home/players/ID8480800/stats/skaterStats/evenTimeOnIce",
        "value": "9:19"
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/home/players/ID8474574/stats/skaterStats/timeOnIce",
        "value": "8:30"
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/home/players/ID8474574/stats/skaterStats/evenTimeOnIce",
        "value": "8:04"
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/home/players/ID8474574/stats/skaterStats/shortHandedTimeOnIce",
        "value": "0:00"
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/home/players/ID8480012/stats/skaterStats/timeOnIce",
        "value": "8:58"
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/home/players/ID8480012/stats/skaterStats/evenTimeOnIce",
        "value": "4:39"
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/home/onIcePlus/2/shiftDuration",
        "value": 317
      }
    ]
  },
  {
    "diff": [
      {
        "op": "replace",
        "path": "/metaData/timeStamp",
        "value": "20210122_040346"
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/away/players/ID8474038/stats/skaterStats/timeOnIce",
        "value": "7:00"
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/away/players/ID8474038/stats/skaterStats/evenTimeOnIce",
        "value": "5:12"
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/away/players/ID8477476/stats/skaterStats/timeOnIce",
        "value": "6:16"
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/away/players/ID8477476/stats/skaterStats/evenTimeOnIce",
        "value": "4:04"
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/away/players/ID8474596/stats/goalieStats/timeOnIce",
        "value": "32:16"
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/away/players/ID8475279/stats/skaterStats/timeOnIce",
        "value": "11:12"
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/away/players/ID8475279/stats/skaterStats/evenTimeOnIce",
        "value": "6:34"
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/away/players/ID8478133/stats/skaterStats/timeOnIce",
        "value": "8:50"
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/away/players/ID8478133/stats/skaterStats/evenTimeOnIce",
        "value": "6:55"
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/away/players/ID8476441/stats/skaterStats/timeOnIce",
        "value": "9:29"
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/away/players/ID8476441/stats/skaterStats/evenTimeOnIce",
        "value": "7:40"
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/home/players/ID8474849/stats/skaterStats/timeOnIce",
        "value": "5:10"
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/home/players/ID8474849/stats/skaterStats/evenTimeOnIce",
        "value": "2:58"
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/home/players/ID8476468/stats/skaterStats/timeOnIce",
        "value": "10:22"
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/home/players/ID8476468/stats/skaterStats/evenTimeOnIce",
        "value": "5:24"
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/home/players/ID8477967/stats/goalieStats/timeOnIce",
        "value": "29:41"
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/home/players/ID8477967/stats/goalieStats/shots",
        "value": 21
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/home/players/ID8477967/stats/goalieStats/saves",
        "value": 18
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/home/players/ID8477967/stats/goalieStats/savePercentage",
        "value": 85.71428571428571
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/home/players/ID8478444/stats/skaterStats/timeOnIce",
        "value": "11:01"
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/home/players/ID8478444/stats/skaterStats/evenTimeOnIce",
        "value": "6:14"
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/home/players/ID8480800/stats/skaterStats/timeOnIce",
        "value": "15:30"
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/home/players/ID8480800/stats/skaterStats/evenTimeOnIce",
        "value": "9:38"
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/home/players/ID8474574/stats/skaterStats/timeOnIce",
        "value": "9:14"
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/home/players/ID8474574/stats/skaterStats/evenTimeOnIce",
        "value": "8:23"
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/home/players/ID8474574/stats/skaterStats/shortHandedTimeOnIce",
        "value": "0:25"
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/home/players/ID8480012/stats/skaterStats/timeOnIce",
        "value": "9:17"
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/home/players/ID8480012/stats/skaterStats/evenTimeOnIce",
        "value": "4:58"
      }
    ]
  },
  {
    "diff": [
      {
        "op": "replace",
        "path": "/metaData/timeStamp",
        "value": "20210122_040407"
      },
      {
        "op": "replace",
        "path": "/liveData/plays/allPlays/115/result/description",
        "value": "Tyler Toffoli (0) Wrist Shot, assists: Joel Armia (3)"
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/away/players/ID8476441/stats/skaterStats/timeOnIce",
        "value": "9:33"
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/away/players/ID8476441/stats/skaterStats/powerPlayTimeOnIce",
        "value": "0:09"
      },
      {
        "op": "remove",
        "path": "/liveData/boxscore/teams/away/onIce/0"
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/away/onIce/1",
        "value": 8475726
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/away/onIce/2",
        "value": 8476469
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/away/onIce/3",
        "value": 8476967
      },
      {
        "op": "add",
        "path": "/liveData/boxscore/teams/away/onIce/5",
        "value": 8481014
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/away/onIcePlus/0/playerId",
        "value": 8474596
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/away/onIcePlus/0/shiftDuration",
        "value": 392
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/away/onIcePlus/0/stamina",
        "value": 33
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/away/onIcePlus/1/playerId",
        "value": 8475726
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/away/onIcePlus/1/shiftDuration",
        "value": 0
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/away/onIcePlus/1/stamina",
        "value": 100
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/away/onIcePlus/2/playerId",
        "value": 8476469
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/away/onIcePlus/2/shiftDuration",
        "value": 0
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/away/onIcePlus/2/stamina",
        "value": 100
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/away/onIcePlus/3/playerId",
        "value": 8476967
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/away/onIcePlus/3/shiftDuration",
        "value": 0
      },
      {
        "op": "remove",
        "path": "/liveData/boxscore/teams/away/onIcePlus/4"
      },
      {
        "op": "add",
        "path": "/liveData/boxscore/teams/away/onIcePlus/5",
        "value": {
          "stamina": 100,
          "shiftDuration": 0,
          "playerId": 8481014
        }
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/home/players/ID8477967/stats/goalieStats/timeOnIce",
        "value": "31:46"
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/home/onIce/1",
        "value": 8476871
      },
      {
        "op": "add",
        "path": "/liveData/boxscore/teams/home/onIce/2",
        "value": 8477500
      },
      {
        "op": "remove",
        "path": "/liveData/boxscore/teams/home/onIce/4"
      },
      {
        "op": "remove",
        "path": "/liveData/boxscore/teams/home/onIce/4"
      },
      {
        "op": "add",
        "path": "/liveData/boxscore/teams/home/onIce/5",
        "value": 8481535
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/home/onIcePlus/1/playerId",
        "value": 8476871
      },
      {
        "op": "replace",
        "path": "/liveData/boxscore/teams/home/onIcePlus/1/shiftDuration",
        "value": 0
      },
      {
        "op": "add",
        "path": "/liveData/boxscore/teams/home/onIcePlus/2",
        "value": {
          "stamina": 100,
          "shiftDuration": 0,
          "playerId": 8477500
        }
      },
      {
        "op": "remove",
        "path": "/liveData/boxscore/teams/home/onIcePlus/4"
      },
      {
        "op": "remove",
        "path": "/liveData/boxscore/teams/home/onIcePlus/4"
      },
      {
        "op": "add",
        "path": "/liveData/boxscore/teams/home/onIcePlus/5",
        "value": {
          "stamina": 100,
          "shiftDuration": 0,
          "playerId": 8481535
        }
      }
    ]
  }
]
`