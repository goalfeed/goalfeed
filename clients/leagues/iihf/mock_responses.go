package iihf

const ActiveGameScoreboard = `
{
  "GameId": "9193",
  "GameNumber": "12",
  "EventId": "503",
  "Gender": "M",
  "Status": "Period 1",
  "GameTime": {
    "TimedGameStatus": "Period 1",
    "PlayTime": "6",
    "Time": "20",
    "PlayTimeMinValue": "0",
    "TimeMaxvalue": "20"
  },
  "IsGameCompleted": false,
  "HomeTeam": {
    "ShortTeamName": "USA",
    "LongTeamName": "United States",
    "Color": "blue"
  },
  "AwayTeam": {
    "ShortTeamName": "CZE",
    "LongTeamName": "Czech Republic",
    "Color": "white"
  },
  "CurrentScore": {
    "Home": "0",
    "Away": "0"
  },
  "Venue": {
    "Code": "EDM",
    "Name": "Rogers Place"
  },
  "EventDateTime": {
    "Day": "29",
    "Month": "DEC",
    "Year": "2020",
    "Weekday": "TUE",
    "Value": "20201229",
    "StartTime": {
      "TimeValue": "1200000000",
      "TimeFormat": "12:00"
    },
    "EndTime": {
      "TimeValue": null,
      "TimeFormat": null
    }
  },
  "Spectators": "",
  "Periods": [
    {
      "Score": {
        "Home": "0",
        "Away": "0"
      },
      "Type": 0,
      "PeriodCode": "1",
      "Period": "1",
      "Displays": {
        "_default": {
          "Title": "Period 1",
          "ShortTitle": "1"
        }
      },
      "Statistics": {
        "FOF_H": "6",
        "FOF_A": "1",
        "SOG_H": "9",
        "SOG_A": "1",
        "TPP_H": "02:00",
        "TPP_A": "00:00",
        "PPG_H": "0",
        "PPG_A": "0",
        "SSG_H": "1",
        "SSG_A": "9",
        "PIM_H": "0",
        "PIM_A": "1",
        "SHG_H": "0",
        "SHG_A": "0"
      },
      "Actions": [
        {
          "Id": null,
          "FullTypeName": "DD.IIHF.Core.Models.DTO.Gamecenter.PlayByPlay.ActionModels.StartRegulationTimeActionModel",
          "Code": "GOL_KPR_IN",
          "TimeOfPlay": null,
          "IsExecutedByHomeTeam": false,
          "ExecutedByShortTeamName": null,
          "ExecutedByLongTeamName": null,
          "ShowExecutedByDot": false,
          "Displays": {
            "_default": {
              "Title": "Start of regulation time",
              "Description": null
            }
          },
          "SituationType": null,
          "SymbolId": "time",
          "Video": null
        },
        {
          "Period": "1",
          "Id": null,
          "FullTypeName": "DD.IIHF.Core.Models.DTO.Gamecenter.PlayByPlay.ActionModels.PeriodStartActionModel",
          "Code": "GOL_KPR_IN",
          "TimeOfPlay": null,
          "IsExecutedByHomeTeam": false,
          "ExecutedByShortTeamName": null,
          "ExecutedByLongTeamName": null,
          "ShowExecutedByDot": false,
          "Displays": {
            "_default": {
              "Title": "Start of period",
              "Description": null
            }
          },
          "SituationType": null,
          "SymbolId": "time",
          "Video": null
        },
        {
          "PenaltyTimeMinutes": "2",
          "PenaltyCode": "Roughing",
          "Athlete": {
            "IH_Athlete_Id": "45172",
            "GG_Athlete_Id": "IHCZE12509200102",
            "NOC_Code": "CZE",
            "Gender": "M",
            "Number": "26",
            "Position": "F",
            "ReportingName": "RASKA Adam",
            "FamilyName": "RASKA",
            "GivenName": "Adam",
            "InitialName": "RASKA A"
          },
          "Id": "1223275",
          "FullTypeName": "DD.IIHF.Core.Models.DTO.Gamecenter.PlayByPlay.ActionModels.PenaltyActionModel",
          "Code": "PTY",
          "TimeOfPlay": "04:04",
          "IsExecutedByHomeTeam": false,
          "ExecutedByShortTeamName": "CZE",
          "ExecutedByLongTeamName": "Czech Republic",
          "ShowExecutedByDot": true,
          "Displays": {
            "_default": {
              "Title": "2min penalty for Czech Republic",
              "Description": "#26 RASKA Adam for Roughing."
            }
          },
          "SituationType": "",
          "SymbolId": "whistle",
          "Video": null
        }
      ],
      "ScoringActions": [
        
      ],
      "TimeLineActions": [
        {
          "PenaltyTimeMinutes": "2",
          "PenaltyCode": "Roughing",
          "Athlete": {
            "IH_Athlete_Id": "45172",
            "GG_Athlete_Id": "IHCZE12509200102",
            "NOC_Code": "CZE",
            "Gender": "M",
            "Number": "26",
            "Position": "F",
            "ReportingName": "RASKA Adam",
            "FamilyName": "RASKA",
            "GivenName": "Adam",
            "InitialName": "RASKA A"
          },
          "Id": "1223275",
          "FullTypeName": "DD.IIHF.Core.Models.DTO.Gamecenter.PlayByPlay.ActionModels.PenaltyActionModel",
          "Code": "PTY",
          "TimeOfPlay": "04:04",
          "IsExecutedByHomeTeam": false,
          "ExecutedByShortTeamName": "CZE",
          "ExecutedByLongTeamName": "Czech Republic",
          "ShowExecutedByDot": true,
          "Displays": {
            "_default": {
              "Title": "2min penalty for Czech Republic",
              "Description": "#26 RASKA Adam for Roughing."
            }
          },
          "SituationType": "",
          "SymbolId": "whistle",
          "Video": null
        }
      ],
      "IceRingActions": [
        
      ]
    },
    {
      "Score": {
        "Home": "0",
        "Away": "0"
      },
      "Type": 3,
      "PeriodCode": "TOT",
      "Period": "TOT",
      "Displays": {
        "_default": {
          "Title": "Total",
          "ShortTitle": "TOT"
        }
      },
      "Statistics": {
        "FOF_H": "6",
        "FOF_A": "1",
        "SOG_H": "9",
        "SOG_A": "1",
        "TPP_H": "02:00",
        "TPP_A": "00:00",
        "PPG_H": "0",
        "PPG_A": "0",
        "SSG_H": "1",
        "SSG_A": "9",
        "PIM_H": "0",
        "PIM_A": "1",
        "SHG_H": "0",
        "SHG_A": "0"
      },
      "Actions": [
        
      ],
      "ScoringActions": [
        
      ],
      "TimeLineActions": [
        
      ],
      "IceRingActions": [
        
      ]
    }
  ],
  "PeriodsGrouped": [
    {
      "PeriodGroupCode": "1",
      "Score": {
        "Home": "0",
        "Away": "0"
      }
    },
    {
      "PeriodGroupCode": "TOT",
      "Score": {
        "Home": "0",
        "Away": "0"
      }
    }
  ]
}
`

// UpcomingGamesSchedule JSON Response for the IIHF schedule. Contains 1 active, 1 final and several upcoming games
const UpcomingGamesSchedule = `
[
  {
    "HomeTeam": {
      "Points": "0",
      "TeamCode": "AUT"
    },
    "GuestTeam": {
      "Points": "0",
      "TeamCode": "SWE"
    },
    "Date": null,
    "GameDateTime": "2020-12-28T16:00:00",
    "GameDateTimeUTC": "2020-12-28T23:00:00Z",
    "EventStatus": "2",
    "Venue": "Rogers Place",
    "VenueCode": "EDM",
    "PhaseId": "PreliminaryRound",
    "Group": "B",
    "Time": null,
    "Status": "FINAL",
    "GameNumber": "10",
    "GameId": "9191",
    "TimeOffset": "GMT-7",
    "TimeOffsetNum": -7
  },
  {
    "HomeTeam": {
      "Points": "0",
      "TeamCode": "SVK"
    },
    "GuestTeam": {
      "Points": "0",
      "TeamCode": "GER"
    },
    "Date": null,
    "GameDateTime": "2020-12-28T19:30:00",
    "GameDateTimeUTC": "2020-12-29T02:30:00Z",
    "EventStatus": "2",
    "Venue": "Rogers Place",
    "VenueCode": "EDM",
    "PhaseId": "PreliminaryRound",
    "Group": "A",
    "Time": null,
    "Status": "FINAL",
    "GameNumber": "11",
    "GameId": "9192",
    "TimeOffset": "GMT-7",
    "TimeOffsetNum": -7
  },
  {
    "HomeTeam": {
      "Points": "0",
      "TeamCode": "USA"
    },
    "GuestTeam": {
      "Points": "0",
      "TeamCode": "CZE"
    },
    "Date": null,
    "GameDateTime": "2020-12-29T12:00:00",
    "GameDateTimeUTC": "2020-12-29T19:00:00Z",
    "EventStatus": "2",
    "Venue": "Rogers Place",
    "VenueCode": "EDM",
    "PhaseId": "PreliminaryRound",
    "Group": "B",
    "Time": null,
    "Status": "LIVE",
    "GameNumber": "12",
    "GameId": "9193",
    "TimeOffset": "GMT-7",
    "TimeOffsetNum": -7
  },
  {
    "HomeTeam": {
      "Points": "0",
      "TeamCode": "CAN"
    },
    "GuestTeam": {
      "Points": "0",
      "TeamCode": "SUI"
    },
    "Date": null,
    "GameDateTime": "2020-12-29T16:00:00",
    "GameDateTimeUTC": "2020-12-29T23:00:00Z",
    "EventStatus": "2",
    "Venue": "Rogers Place",
    "VenueCode": "EDM",
    "PhaseId": "PreliminaryRound",
    "Group": "A",
    "Time": null,
    "Status": "UPCOMING",
    "GameNumber": "13",
    "GameId": "9194",
    "TimeOffset": "GMT-7",
    "TimeOffsetNum": -7
  },
  {
    "HomeTeam": {
      "Points": "0",
      "TeamCode": "AUT"
    },
    "GuestTeam": {
      "Points": "0",
      "TeamCode": "RUS"
    },
    "Date": null,
    "GameDateTime": "2020-12-29T19:30:00",
    "GameDateTimeUTC": "2020-12-30T02:30:00Z",
    "EventStatus": "2",
    "Venue": "Rogers Place",
    "VenueCode": "EDM",
    "PhaseId": "PreliminaryRound",
    "Group": "B",
    "Time": null,
    "Status": "UPCOMING",
    "GameNumber": "14",
    "GameId": "9195",
    "TimeOffset": "GMT-7",
    "TimeOffsetNum": -7
  },
  {
    "HomeTeam": {
      "Points": "0",
      "TeamCode": "FIN"
    },
    "GuestTeam": {
      "Points": "0",
      "TeamCode": "SVK"
    },
    "Date": null,
    "GameDateTime": "2020-12-30T12:00:00",
    "GameDateTimeUTC": "2020-12-30T19:00:00Z",
    "EventStatus": "2",
    "Venue": "Rogers Place",
    "VenueCode": "EDM",
    "PhaseId": "PreliminaryRound",
    "Group": "A",
    "Time": null,
    "Status": "UPCOMING",
    "GameNumber": "15",
    "GameId": "9196",
    "TimeOffset": "GMT-7",
    "TimeOffsetNum": -7
  },
  {
    "HomeTeam": {
      "Points": "0",
      "TeamCode": "SUI"
    },
    "GuestTeam": {
      "Points": "0",
      "TeamCode": "GER"
    },
    "Date": null,
    "GameDateTime": "2020-12-30T16:00:00",
    "GameDateTimeUTC": "2020-12-30T23:00:00Z",
    "EventStatus": "2",
    "Venue": "Rogers Place",
    "VenueCode": "EDM",
    "PhaseId": "PreliminaryRound",
    "Group": "A",
    "Time": null,
    "Status": "UPCOMING",
    "GameNumber": "16",
    "GameId": "9197",
    "TimeOffset": "GMT-7",
    "TimeOffsetNum": -7
  },
  {
    "HomeTeam": {
      "Points": "0",
      "TeamCode": "RUS"
    },
    "GuestTeam": {
      "Points": "0",
      "TeamCode": "SWE"
    },
    "Date": null,
    "GameDateTime": "2020-12-30T19:30:00",
    "GameDateTimeUTC": "2020-12-31T02:30:00Z",
    "EventStatus": "2",
    "Venue": "Rogers Place",
    "VenueCode": "EDM",
    "PhaseId": "PreliminaryRound",
    "Group": "B",
    "Time": null,
    "Status": "UPCOMING",
    "GameNumber": "17",
    "GameId": "9198",
    "TimeOffset": "GMT-7",
    "TimeOffsetNum": -7
  },
  {
    "HomeTeam": {
      "Points": "0",
      "TeamCode": "CZE"
    },
    "GuestTeam": {
      "Points": "0",
      "TeamCode": "AUT"
    },
    "Date": null,
    "GameDateTime": "2020-12-31T12:00:00",
    "GameDateTimeUTC": "2020-12-31T19:00:00Z",
    "EventStatus": "2",
    "Venue": "Rogers Place",
    "VenueCode": "EDM",
    "PhaseId": "PreliminaryRound",
    "Group": "B",
    "Time": null,
    "Status": "UPCOMING",
    "GameNumber": "18",
    "GameId": "9199",
    "TimeOffset": "GMT-7",
    "TimeOffsetNum": -7
  },
  {
    "HomeTeam": {
      "Points": "0",
      "TeamCode": "CAN"
    },
    "GuestTeam": {
      "Points": "0",
      "TeamCode": "FIN"
    },
    "Date": null,
    "GameDateTime": "2020-12-31T16:00:00",
    "GameDateTimeUTC": "2020-12-31T23:00:00Z",
    "EventStatus": "2",
    "Venue": "Rogers Place",
    "VenueCode": "EDM",
    "PhaseId": "PreliminaryRound",
    "Group": "A",
    "Time": null,
    "Status": "UPCOMING",
    "GameNumber": "19",
    "GameId": "9200",
    "TimeOffset": "GMT-7",
    "TimeOffsetNum": -7
  },
  {
    "HomeTeam": {
      "Points": "0",
      "TeamCode": "SWE"
    },
    "GuestTeam": {
      "Points": "0",
      "TeamCode": "USA"
    },
    "Date": null,
    "GameDateTime": "2020-12-31T19:30:00",
    "GameDateTimeUTC": "2021-01-01T02:30:00Z",
    "EventStatus": "2",
    "Venue": "Rogers Place",
    "VenueCode": "EDM",
    "PhaseId": "PreliminaryRound",
    "Group": "B",
    "Time": null,
    "Status": "UPCOMING",
    "GameNumber": "20",
    "GameId": "9201",
    "TimeOffset": "GMT-7",
    "TimeOffsetNum": -7
  },
  {
    "HomeTeam": {
      "Points": "0",
      "TeamCode": "QF"
    },
    "GuestTeam": {
      "Points": "0",
      "TeamCode": "QF"
    },
    "Date": null,
    "GameDateTime": "2021-01-02T10:00:00",
    "GameDateTimeUTC": "2021-01-02T17:00:00Z",
    "EventStatus": "2",
    "Venue": "Rogers Place",
    "VenueCode": "EDM",
    "PhaseId": "Quarterfinals",
    "Group": "",
    "Time": null,
    "Status": "UPCOMING",
    "GameNumber": "21",
    "GameId": "9202",
    "TimeOffset": "GMT-7",
    "TimeOffsetNum": -7
  },
  {
    "HomeTeam": {
      "Points": "0",
      "TeamCode": "QF"
    },
    "GuestTeam": {
      "Points": "0",
      "TeamCode": "QF"
    },
    "Date": null,
    "GameDateTime": "2021-01-02T13:30:00",
    "GameDateTimeUTC": "2021-01-02T20:30:00Z",
    "EventStatus": "2",
    "Venue": "Rogers Place",
    "VenueCode": "EDM",
    "PhaseId": "Quarterfinals",
    "Group": "",
    "Time": null,
    "Status": "UPCOMING",
    "GameNumber": "22",
    "GameId": "9203",
    "TimeOffset": "GMT-7",
    "TimeOffsetNum": -7
  },
  {
    "HomeTeam": {
      "Points": "0",
      "TeamCode": "QF"
    },
    "GuestTeam": {
      "Points": "0",
      "TeamCode": "QF"
    },
    "Date": null,
    "GameDateTime": "2021-01-02T17:00:00",
    "GameDateTimeUTC": "2021-01-03T00:00:00Z",
    "EventStatus": "2",
    "Venue": "Rogers Place",
    "VenueCode": "EDM",
    "PhaseId": "Quarterfinals",
    "Group": "",
    "Time": null,
    "Status": "UPCOMING",
    "GameNumber": "23",
    "GameId": "9204",
    "TimeOffset": "GMT-7",
    "TimeOffsetNum": -7
  },
  {
    "HomeTeam": {
      "Points": "0",
      "TeamCode": "QF"
    },
    "GuestTeam": {
      "Points": "0",
      "TeamCode": "QF"
    },
    "Date": null,
    "GameDateTime": "2021-01-02T20:30:00",
    "GameDateTimeUTC": "2021-01-03T03:30:00Z",
    "EventStatus": "2",
    "Venue": "Rogers Place",
    "VenueCode": "EDM",
    "PhaseId": "Quarterfinals",
    "Group": "",
    "Time": null,
    "Status": "UPCOMING",
    "GameNumber": "24",
    "GameId": "9205",
    "TimeOffset": "GMT-7",
    "TimeOffsetNum": -7
  },
  {
    "HomeTeam": {
      "Points": "0",
      "TeamCode": "SF"
    },
    "GuestTeam": {
      "Points": "0",
      "TeamCode": "SF"
    },
    "Date": null,
    "GameDateTime": "2021-01-04T16:00:00",
    "GameDateTimeUTC": "2021-01-04T23:00:00Z",
    "EventStatus": "2",
    "Venue": "Rogers Place",
    "VenueCode": "EDM",
    "PhaseId": "Semifinals",
    "Group": "",
    "Time": null,
    "Status": "UPCOMING",
    "GameNumber": "25",
    "GameId": "9206",
    "TimeOffset": "GMT-7",
    "TimeOffsetNum": -7
  },
  {
    "HomeTeam": {
      "Points": "0",
      "TeamCode": "SF"
    },
    "GuestTeam": {
      "Points": "0",
      "TeamCode": "SF"
    },
    "Date": null,
    "GameDateTime": "2021-01-04T19:30:00",
    "GameDateTimeUTC": "2021-01-05T02:30:00Z",
    "EventStatus": "2",
    "Venue": "Rogers Place",
    "VenueCode": "EDM",
    "PhaseId": "Semifinals",
    "Group": "",
    "Time": null,
    "Status": "UPCOMING",
    "GameNumber": "26",
    "GameId": "9207",
    "TimeOffset": "GMT-7",
    "TimeOffsetNum": -7
  },
  {
    "HomeTeam": {
      "Points": "0",
      "TeamCode": "L(SF)"
    },
    "GuestTeam": {
      "Points": "0",
      "TeamCode": "L(SF)"
    },
    "Date": null,
    "GameDateTime": "2021-01-05T15:30:00",
    "GameDateTimeUTC": "2021-01-05T22:30:00Z",
    "EventStatus": "2",
    "Venue": "Rogers Place",
    "VenueCode": "EDM",
    "PhaseId": "BronzeMedalGame",
    "Group": "",
    "Time": null,
    "Status": "UPCOMING",
    "GameNumber": "27",
    "GameId": "9208",
    "TimeOffset": "GMT-7",
    "TimeOffsetNum": -7
  },
  {
    "HomeTeam": {
      "Points": "0",
      "TeamCode": "W(SF)"
    },
    "GuestTeam": {
      "Points": "0",
      "TeamCode": "W(SF)"
    },
    "Date": null,
    "GameDateTime": "2021-01-05T19:30:00",
    "GameDateTimeUTC": "2021-01-06T02:30:00Z",
    "EventStatus": "2",
    "Venue": "Rogers Place",
    "VenueCode": "EDM",
    "PhaseId": "GoldMedalGame",
    "Group": "",
    "Time": null,
    "Status": "UPCOMING",
    "GameNumber": "28",
    "GameId": "9209",
    "TimeOffset": "GMT-7",
    "TimeOffsetNum": -7
  }
]
`
