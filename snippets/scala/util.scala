import java.io.File

import scala.util.matching.Regex


object MinishUtil {
  def findLatest(dir: String, fileRegExp: Regex): Option[String] = {
    var modifiedTime: Long = 0
    var newestFile = ""

    val path = new File(dir)
    for (file <- path.listFiles) {
      file.getName match {
        case fileRegExp(name) => {
          if (file.lastModified() > modifiedTime) {
            newestFile = name
            modifiedTime = file.lastModified()
          }
        }
        case ignore => {}
      }
    }

    newestFile != "" match {
      case true => Some(newestFile)
      case _ => {
        None
      }
    }
  }

  def main(args: Array[String]) = {
    findLatest("../../dist", """(^\S+-example\.min\.js$)""".r) match {
      case Some(latest) => println("Latest: " + latest)
      case failed => {
        println("Failed to find latest: " + failed)
      }
    }
  }
}