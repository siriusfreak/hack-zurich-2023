import React, { useState } from 'react';
import { Container, TextField, Button, List, ListItem, ListItemText, Divider, MenuItem, Menu, IconButton } from '@mui/material';
import SendIcon from '@mui/icons-material/Send';
import sikaLogo from './assets/LogoSika.png';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import Flag from 'react-country-flag';

const styles = {
  chatContainer: {
    padding: '50px',
    border: '1px solid #ccc',
    borderRadius: '10px',
    background: '#ffffff',
    height: '100vh',
    display: 'flex',
    flexDirection: 'column',
    justifyContent: 'space-between',
    margin: '0 auto',
    maxWidth: '800px',
  },
  chatHeader: {
    backgroundColor: '#ffffff',
    borderBottom: '1px solid #ccc',
    padding: '10px',
    fontSize: '20px',
    fontWeight: 'bold',
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
  },  
  sikaLogoImage: {
    width: '50px',
    height: 'auto',
  },
  messageList: {
    display: 'flex',
    flexDirection: 'column',
    flexGrow: 1,
    overflow: 'auto',
  },
  messageBubble: (isOwnMessage) => ({
    alignSelf: isOwnMessage ? 'flex-end' : 'flex-start',
    maxWidth: '50%',
    padding: '10px 20px',
    borderRadius: '20px',
    margin: '10px',
    backgroundColor: '#f5f5f5',
    textAlign: isOwnMessage ? 'right' : 'left',
  }),
  inputContainer: {
    display: 'flex',
    padding: '10px',
    backgroundColor: '#f0f0f0',
    borderRadius: '10px',
    position: 'relative',
  },
  inputField: {
    flexGrow: 1,
    borderRadius: '10px',
    overflow: 'hidden',
    marginRight: '10px',
  },
  sendButton: {
    position: 'absolute',
    right: '10px',
    top: '50%',
    transform: 'translateY(-50%)',
    borderRadius: '50%',
    height: '40px',
    width: '40px',
    boxShadow: 'none',
    '&:hover': {
      backgroundColor: 'rgba(0, 0, 0, 0.04)',
    },
  },
  languageSelector: {
    display: 'flex',
    alignItems: 'center',
  },
};

function Chat() {
  const [messages, setMessages] = useState([]);
  const [inputValue, setInputValue] = useState('');
  const [currentUser, setCurrentUser] = useState('You');
  const [selectedLanguage, setSelectedLanguage] = useState('english');
  const [languageAnchorEl, setLanguageAnchorEl] = useState(null);

  const chatID = 111;

  const handleSend = () => {
    if (inputValue.trim()) {
      const newMessage = { text: inputValue, sender: currentUser };
      setMessages((prevMessages) => [...prevMessages, newMessage]);
      setInputValue('');

      fetch(`http://localhost:8080/chat/${chatID}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ message: inputValue, language: selectedLanguage }),
      })
        .then((response) => response.json())
        .then((data) => {
          setMessages((prevMessages) => [
            ...prevMessages,
            { text: data.response, sender: currentUser === 'You' ? 'Assistant' : 'You' },
          ]);
        })
        .catch((error) => console.error('Error:', error));

      setCurrentUser((prevUser) => (prevUser === 'You' ? 'Assistant' : 'You'));
    }
  };

  const handleKeyPress = (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  const handleLanguageClick = (event) => {
    setLanguageAnchorEl(event.currentTarget);
  };

  const handleLanguageClose = () => {
    setLanguageAnchorEl(null);
  };

  const handleLanguageChange = (language) => {
    setSelectedLanguage(language);
    setLanguageAnchorEl(null);
  };

  return (
    <Container sx={styles.chatContainer}>
      <div sx={styles.chatHeader}>
        <div style={{ display: 'flex', alignItems: 'center' }}> 
            <img src={sikaLogo} alt="Sika Logo" style={styles.sikaLogoImage} />
        </div>

        <div style={{ display: 'flex', alignItems: 'center' }}>
        <IconButton
            aria-controls="language-menu"
            aria-haspopup="true"
            onClick={handleLanguageClick}
            sx={styles.languageSelector}
            >
            <Flag countryCode={getCountryCode(selectedLanguage)} style={{ marginRight: '8px' }} />
            <ExpandMoreIcon />
        </IconButton>
        <Menu
          id="language-menu"
          anchorEl={languageAnchorEl}
          keepMounted
          open={Boolean(languageAnchorEl)}
          onClose={handleLanguageClose}
        >
          <MenuItem onClick={() => handleLanguageChange('english')}>
            <Flag countryCode="GB" style={{ marginRight: '8px' }} />
            English
          </MenuItem>
          <MenuItem onClick={() => handleLanguageChange('german')}>
            <Flag countryCode="DE" style={{ marginRight: '8px' }} />
            German
          </MenuItem>
          <MenuItem onClick={() => handleLanguageChange('portuguese')}>
            <Flag countryCode="PT" style={{ marginRight: '8px' }} />
            Portuguese
          </MenuItem>
          <MenuItem onClick={() => handleLanguageChange('spanish')}>
            <Flag countryCode="ES" style={{ marginRight: '8px' }} />
            Spanish
          </MenuItem>
          <MenuItem onClick={() => handleLanguageChange('chinese')}>
            <Flag countryCode="CN" style={{ marginRight: '8px' }} />
            Chinese
          </MenuItem>
        </Menu>
        </div>
      </div>

      <Divider />
      <List sx={styles.messageList}>
        {messages.map((message, index) => (
          <ListItem key={index} sx={styles.messageBubble(message.sender === 'You')}>
            <ListItemText primary={message.sender} secondary={message.text} />
          </ListItem>
        ))}
      </List>
      <div sx={styles.inputContainer}>
        <TextField
          sx={{
            ...styles.inputField,
            '& .MuiOutlinedInput-root': {
              '&.Mui-focused fieldset': {
                borderColor: 'red',
              },
              '& fieldset': {
                borderColor: 'red',
              },
            },
          }}
          variant="outlined"
          value={inputValue}
          onChange={(e) => setInputValue(e.target.value)}
          onKeyPress={handleKeyPress}
          fullWidth
          placeholder="Write a message ..."
          InputProps={{
            endAdornment: (
              <Button sx={styles.sendButton} color="primary" onClick={handleSend}>
                <SendIcon sx={{ color: 'red' }} />
              </Button>
            ),
            style: { backgroundColor: 'rgba(255, 255, 255, 0.5)' },
          }}
        />
      </div>
    </Container>
  );
}

function getCountryCode(language) {
    switch (language) {
      case 'english':
        return 'GB';
      case 'german':
        return 'DE';
      case 'portuguese':
        return 'PT';
      case 'spanish':
        return 'ES';
      case 'chinese':
        return 'CN';
      default:
        return 'US'; // Используйте код страны по умолчанию, если язык не распознан
    }
  }  

export default Chat;
