const K = 30;
const INITIAL_RATING = 500;
const NO_OF_POINT_REDUCTION_MATCHES = 25;
const NO_OF_POINT_REDUCTION = 10;

function calculateEloAndSendToChat() {
  const sheet = SpreadsheetApp.getActiveSpreadsheet().getSheetByName('Matches');
  const data = sheet.getRange(2, 2, sheet.getLastRow() - 1, 4).getValues(); // Get data from columns C, D, E starting from row 2
  
  let ratings = {};
  let lastDate = ""
  let lastPlayedMatch = {}
  
  // Iterate through the rows of the sheet
  for (let i = 0; i < data.length; i++) {
    if (data[i][0]) {
      lastDate = data[i][0]
    }
    const player1 = data[i][1]; // Player 1 in column C
    const player2 = data[i][2]; // Player 2 in column D
    const winner = data[i][3];  // Winner in column E
    
    let player1Rating = ratings[player1] || INITIAL_RATING;
    let player2Rating = ratings[player2] || INITIAL_RATING;
    
    const p1 = 1 / (1 + Math.pow(10, (player2Rating - player1Rating) / 400));
    const p2 = 1 / (1 + Math.pow(10, (player1Rating - player2Rating) / 400));
    
    let player1Actual = 0;
    let player2Actual = 0;
    
    if (winner === player1) {
      player1Actual = 1;
      player2Actual = 0;
    } else if (winner === player2) {
      player2Actual = 1;
      player1Actual = 0;
    }
    
    ratings[player1] = player1Rating + K * (player1Actual - p1);
    ratings[player2] = player2Rating + K * (player2Actual - p2);
    if (lastPlayedMatch[player1] < i-NO_OF_POINT_REDUCTION_MATCHES) {
      ratings[player1] = ratings[player1] - NO_OF_POINT_REDUCTION
    }
    if (lastPlayedMatch[player2] < i-NO_OF_POINT_REDUCTION_MATCHES) {
      ratings[player2] = ratings[player2] - NO_OF_POINT_REDUCTION
    }
    lastPlayedMatch[player1]=i
    lastPlayedMatch[player2]=i
  }
  
  for (let player in ratings) {
    if (lastPlayedMatch[player] < data.length-NO_OF_POINT_REDUCTION_MATCHES) {
      ratings[player] = ratings[player] - NO_OF_POINT_REDUCTION
    }
  }

  // Sort the players by their rating in descending order
  let players = Object.keys(ratings).sort((a, b) => ratings[b] - ratings[a]);
  
  // Prepare the message to send
  let msg = `Final ratings as of ${new Date().toISOString().slice(0, 10)}:`;
  
  players.forEach(player => {
    msg += `\n${player}: ${ratings[player].toFixed(2)}`;
  });
  
  // Send the message to Google Chat
  const webhookUrl = ""
  if (isYesterday(lastDate) ) {
    sendMessageToGoogleChat(webhookUrl, msg);
  } else {
    Logger.log("skipping sending chat")
  }
}

// Helper function to send the message to Google Chat via Webhook
function sendMessageToGoogleChat(webhookUrl, message) {
  const payload = JSON.stringify({text: message});
  
  const options = {
    method: 'POST',
    contentType: 'application/json',
    payload: payload
  };
  
  const response = UrlFetchApp.fetch(webhookUrl, options);
  
  if (response.getResponseCode() !== 200) {
    Logger.log(`Error: Received non-200 response: ${response.getResponseCode()}`);
  } else {
    Logger.log("Message successfully sent to Google Chat.");
  }
}

function isYesterday(lastDate) {
  // Create a new Date object for yesterday
  var yesterday = new Date();
  yesterday.setDate(yesterday.getDate() - 1);

  // Convert lastDate to a Date object if it's not already
  var lastDateObj = new Date(lastDate);
  // Compare only the date part (ignore time)
  return lastDateObj.toDateString() === yesterday.toDateString();
}
