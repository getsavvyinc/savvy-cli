import type React from "react"
import { useState, useEffect } from "react"
import { Button } from "@src/components/ui/button"
import { Dialog, DialogFooter, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@src/components/ui/dialog"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@src/components/ui/tabs"
import { Switch } from "@src/components/ui/switch"
import { Label } from "@src/components/ui/label"
import { Copy } from "lucide-react"
import { toast } from 'sonner';

interface HistoryItem extends chrome.history.HistoryItem {
  isSelected?: boolean
}

interface CopyURLsProps {
  selectedItems: HistoryItem[]
}

const extensionURL = 'https://chromewebstore.google.com/detail/savvy/jocphfjphhfbdccjfjjnbcnejmbojjlh'

export const CopyURLs: React.FC<CopyURLsProps> = ({ selectedItems }) => {
  const [includeTimestamp, setIncludeTimestamp] = useState(false)
  const [isOpen, setIsOpen] = useState(false)
  const [copiedText, setCopiedText] = useState<boolean>(false);
  const [copiedMd, setCopiedMd] = useState<boolean>(false);


  const getRawText = () => {
    const prependSavvy = "Here's the list of URLs (generated using Savvy):\n\n"
    const prependSavvySingular = "Here's the URL (generated using Savvy):\n\n"
    const text =  selectedItems.map((item) => "- " + item.url).join("\n")
    if (selectedItems.length === 1) {
      return prependSavvySingular + text
    }
    return prependSavvy + text
  }

  const getMarkdownText = () => {
    const prependSavvy = `Here's the list of URLs (generated using [Savvy](${extensionURL}))`
    const prependSavvySingular = `Here's the URL (generated using [Savvy](${extensionURL}))`
    const mdText =  selectedItems
      .map((item) => {
        const title = item.title || new URL(item.url || "").hostname
        const timestamp = includeTimestamp ? `${new Date(item.lastVisitTime!).toLocaleString()} ` : `` 
        return `- ${timestamp}[${title}](${item.url})`
      })
      .join("\n")

      if (selectedItems.length === 1) {
        return `${prependSavvySingular}\n\n${mdText}`
      }
      return `${prependSavvy}\n\n${mdText}`
  }

  const copyMdToClipboard = (text: string) => {
    navigator.clipboard
      .writeText(text)
      .then(() => {
        setCopiedMd(true)
      })
      .catch((err) => {
      toast.error("Failed to copy text to clipboard", {
        duration: 3000,
        position: 'top-right',
        closeButton: true,
      })
      console.error(err)
      })
  }
  
  const copyTextToClipboard = (text: string) => {
    navigator.clipboard.writeText(text).then(() => {
      setCopiedText(true)
    }).catch((err) => {
      toast.error("Failed to copy text to clipboard", {
        duration: 3000,
        position: 'top-right',
        closeButton: true,
      })
      console.error(err)
    })
  }


   useEffect(() => {
        if (copiedText) {
            const timer = window.setTimeout(() => {
                setCopiedText(false);
            }, 3000);

            // Clear the timeout if the component is unmounted before the timer is up
            return () => window.clearTimeout(timer);
        }
    }, [copiedText]);

   useEffect(() => {
        if (copiedMd) {
            const timer = window.setTimeout(() => {
                setCopiedMd(false);
            }, 3000);

            // Clear the timeout if the component is unmounted before the timer is up
            return () => window.clearTimeout(timer);
        }
    }, [copiedMd]);

  return (
    <>
      <Dialog open={isOpen} onOpenChange={setIsOpen} modal={false}>
        <DialogTrigger asChild className="w-full">
          <Button variant="outline" className="mb-2 w-full" disabled={selectedItems.length === 0}>
            <Copy className="mr-2 size-4" />
            Copy 
          </Button>
        </DialogTrigger>
        <DialogContent className="flex w-full flex-col gap-4 overflow-scroll">
          <DialogHeader>
            <DialogTitle>Export History</DialogTitle>
          </DialogHeader>
          <Tabs defaultValue="rawtext" className="w-full">
            <TabsList className="grid w-full grid-cols-2">
              <TabsTrigger value="rawtext">Text</TabsTrigger>
              <TabsTrigger value="markdown">Markdown</TabsTrigger>
            </TabsList>
            <TabsContent value="rawtext" className="flex w-full flex-col gap-8">
              <div className="w-full overflow-scroll rounded-md border p-4">
                <pre className="font-monospace max-h-[66vh] min-h-[33vh] font-thin">{getRawText()}</pre>
              </div>
              <DialogFooter>
              <Button onClick={() => copyTextToClipboard(getRawText())} className="mt-4 w-full text-white">
                { copiedText ? "Copied!" : "Copy History" }
              </Button>
              </DialogFooter>
            </TabsContent>
            <TabsContent value="markdown">
              <div className="mb-4 flex items-center space-x-2">
                <Switch id="timestamp-mode" checked={includeTimestamp} onCheckedChange={setIncludeTimestamp} />
                <Label htmlFor="timestamp-mode">
                  {selectedItems.length > 1 ? "Include Timestamps" : "Include Timestamp"}
                </Label>
              </div>
              <div className="w-full overflow-scroll rounded-md border p-4">
                <pre className="font-monospace max-h-[66vh] min-h-[33vh] font-thin">{getMarkdownText()}</pre>
              </div>
              <DialogFooter>
              <Button onClick={() => copyMdToClipboard(getMarkdownText())} className="mt-4 w-full text-white">
                { copiedMd ? "Copied!" : "Copy Markdown" }
              </Button>
              </DialogFooter>
            </TabsContent>
          </Tabs>
        </DialogContent>
      </Dialog>
    </>
  )
}
