<?php
$ans = $_SERVER['QUERY_STRING'];
sleep(1);
$ans = strtoupper ( $ans );
switch ($ans) {
    case "DOCKS": //In Space, No One Can Hear You Rock Out
        header("Location: http://puzzlehunt.club.cc.cmu.edu/assets/internal/DOCKS.gif");
        exit;
        break;
    case "FORGET": //Museum of Wikipedia
        header("Location: http://puzzlehunt.club.cc.cmu.edu/assets/internal/FORGET.gif");
        exit;
        break;
    case "HYPERBOLICPARABOLOID": //This is a Crossword (Really!) 
        header("Location: http://puzzlehunt.club.cc.cmu.edu/assets/internal/HYPERBOLICPARABOLOID.mp3");
        exit;
        break;
    case "PARENTAL": //Old Games 
        header("Location: http://puzzlehunt.club.cc.cmu.edu/assets/internal/PARENTAL.gif");
        exit;
        break;
    case "DRIGO": //Downwords 
        header("Location: http://puzzlehunt.club.cc.cmu.edu/assets/internal/DRIGO.gif");
        exit;
        break;
    case "PRAETORIANGUARD": //Intersections 
        header("Location: http://puzzlehunt.club.cc.cmu.edu/assets/internal/PRAETORIANGUARD.mp3");
        exit;
        break;
    case "CREATOR": //Sonobe Conundrum  
        header("Location: http://puzzlehunt.club.cc.cmu.edu/assets/internal/CREATOR.gif");
        exit;
        break;
    case "MCDONALDS": //Nonverbal Communication
        header("Location: http://puzzlehunt.club.cc.cmu.edu/assets/internal/MCDONALDS.gif");
        exit;
        break;
    case "SPACE": //Not a Crossword (Really!) 
        header("Location: http://puzzlehunt.club.cc.cmu.edu/assets/internal/SPACE.gif");
        exit;
        break;
    case "PARADOX": //Pop Quiz  
        header("Location: http://puzzlehunt.club.cc.cmu.edu/assets/internal/PARADOX.gif");
        exit;
        break;
    case "NINES": //Unhelpful Indices 
        header("Location: http://puzzlehunt.club.cc.cmu.edu/assets/internal/NINES.mp3");
        exit;
        break;
    case "GANONDORF": //Links 
        header("Location: http://puzzlehunt.club.cc.cmu.edu/assets/internal/GANONDORF.gif");
        exit;
        break;
    case "NEGATIONS": //A Lightheaded Chapter 
        header("Location: http://puzzlehunt.club.cc.cmu.edu/assets/internal/NEGATIONS.mp3");
        exit;
        break;
    case "SINE": //Sign Error
        header("Location: http://puzzlehunt.club.cc.cmu.edu/assets/internal/SINE.gif");
        exit;
        break;
}
?>